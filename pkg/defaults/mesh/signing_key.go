package mesh

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func ensureDataplaneTokenSigningKey(resManager manager.ResourceManager, meshName string) (created bool, err error) {
<<<<<<< HEAD
	return ensureSigningKeyForPrefix(resManager, meshName, issuer.DataplaneTokenPrefix)
}

func ensureEnvoyAdminClientSigningKey(resManager manager.ResourceManager, meshName string) (created bool, err error) {
	return ensureSigningKeyForPrefix(resManager, meshName, issuer.EnvoyAdminClientTokenPrefix)
}

func ensureSigningKeyForPrefix(resManager manager.ResourceManager, meshName, prefix string) (created bool, err error) {
	signingKey, err := issuer.CreateSigningKey()
	if err != nil {
		return false, errors.Wrap(err, "could not create a signing key")
	}
	key := issuer.SigningKeyResourceKey(prefix, meshName)
	err = resManager.Get(context.Background(), signingKey, core_store.GetBy(key))
=======
	ctx := context.Background()
	signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, issuer.DataplaneTokenSigningKeyPrefix(meshName), meshName)
	_, _, err = signingKeyManager.GetLatestSigningKey(ctx)
>>>>>>> 8708885a (feat(kuma-cp) consolidate tokens logic to support expiration, rotation, revocation and RSA256 (#3376))
	if err == nil {
		return false, nil
	}
	if err != nil && !tokens.IsSigningKeyNotFound(err) {
		return false, err
	}
	if err := signingKeyManager.CreateDefaultSigningKey(ctx); err != nil {
		return false, err
	}
	return true, nil
}
