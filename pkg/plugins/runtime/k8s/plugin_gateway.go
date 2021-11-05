package k8s

import (
	"fmt"

	"github.com/pkg/errors"
	kube_ctrl "sigs.k8s.io/controller-runtime"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	gatewayapi_controllers "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks/injector"
)

func addGatewayReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	// If we haven't registered our type, we're not reconciling gatewayapi
	// objects.
	if _, err := registry.Global().NewObject(&mesh_proto.Gateway{}); err != nil {
		var unknownTypeError *registry.UnknownTypeError
		if errors.As(err, &unknownTypeError) {
			return nil
		}
	}

	cpURL := fmt.Sprintf("https://%s.%s:%d", rt.Config().Runtime.Kubernetes.ControlPlaneServiceName, rt.Config().Store.Kubernetes.SystemNamespace, rt.Config().DpServer.Port)

	kumaInjector, err := injector.New(
		rt.Config().Runtime.Kubernetes.Injector,
		cpURL,
		mgr.GetClient(),
		converter,
	)
	if err != nil {
		return errors.Wrap(err, "could not create Kuma injector")
	}

	gatewayAPIGatewayReconciler := &gatewayapi_controllers.GatewayReconciler{
		Client:          mgr.GetClient(),
		Reader:          mgr.GetAPIReader(),
		Log:             core.Log.WithName("controllers").WithName("gatewayapi").WithName("Gateway"),
		Scheme:          mgr.GetScheme(),
		Converter:       converter,
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		Injector:        *kumaInjector,
		ResourceManager: rt.ResourceManager(),
	}
	if err := gatewayAPIGatewayReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API Gateway reconciler")
	}

	gatewayAPIHTTPRouteReconciler := &gatewayapi_controllers.HTTPRouteReconciler{
		Client:          mgr.GetClient(),
		Reader:          mgr.GetAPIReader(),
		Log:             core.Log.WithName("controllers").WithName("gatewayapi").WithName("HTTPRoute"),
		Scheme:          mgr.GetScheme(),
		Converter:       converter,
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceManager: rt.ResourceManager(),
	}
	if err := gatewayAPIHTTPRouteReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API HTTPRoute reconciler")
	}

	return nil
}
