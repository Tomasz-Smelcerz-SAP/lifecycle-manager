package zerodw

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	apicorev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	gatewaySecretName = "gateway-secret"
)

var (
	CABundleNotFound = errors.New("CA bundle secret not found")
)

type CABundleFinder interface {
	FindCABundleSecret() (*apicorev1.Secret, error)
}

type gatewaySecretHandler struct {
	*secretManager
	CABundleFinder
}

func NewGatewaySecretHandler(caBundleFinder CABundleFinder, kcpClient client.Client, log logr.Logger) *gatewaySecretHandler {
	return &gatewaySecretHandler{
		secretManager: &secretManager{
			kcpClient: kcpClient,
			log:       log,
		},
		CABundleFinder: caBundleFinder,
	}
}

func (gsh *gatewaySecretHandler) ManageGatewaySecret() error {

	caBundle, err := gsh.FindCABundleSecret()

	if isNotFound(err) {
		// ca-bundle secret does not exist, we can't configure gateway without it
		gsh.log.Error(CABundleNotFound, "gatewaySecretHandler")
		return CABundleNotFound
	}

	secret, err := gsh.findGatewaySecret()

	if isNotFound(err) {
		// gateway secret does not exist
		return gsh.handleNonExisting(caBundle)
	}
	if err != nil {
		return err
	}

	// gateway secret exists
	return gsh.handleExisting(caBundle, secret)
}

func (gsh *gatewaySecretHandler) handleNonExisting(caBundle *apicorev1.Secret) error {
	// create gateway secret
	gwSecret := gsh.newGatewaySecret(caBundle)
	err := gsh.create(context.TODO(), gwSecret)
	if err == nil {
		gsh.log.Info("created the gateway secret", "reason", "gateway secret does not exist")
	}
	return err
}

func (gsh *gatewaySecretHandler) handleExisting(caBundle *apicorev1.Secret, gwSecret *apicorev1.Secret) error {

	// skip the update only if the creation time of caBundle is older than the gateway Secret timestamp
	doUpdate := true

	lastModifiedAtValue, ok := gwSecret.Annotations[lastModifiedAtAnnotation]
	if ok {
		gwSecretLastModifiedAt, err := time.Parse(time.RFC3339, lastModifiedAtValue)
		if err == nil && caBundle.CreationTimestamp.Time.Before(gwSecretLastModifiedAt) {
			doUpdate = false
		}
	}

	// update gateway secret if creation time of caBundle is newer than the gateway Secret
	if doUpdate {
		//create the log entry again that describes the reason for the Update
		gwSecret.Data["tls.crt"] = caBundle.Data["root.tls.crt"]
		gwSecret.Data["tls.key"] = caBundle.Data["root.tls.key"]
		gwSecret.Data["ca.crt"] = joinCACertsFromBundle(caBundle)
		err := gsh.update(context.TODO(), gwSecret)
		if err == nil {
			gsh.log.Info("updated the gateway secret", "reason", "CA-Bundle is more recent than the gateway secret")
		}
	}

	return nil
}

func (gsh *gatewaySecretHandler) findGatewaySecret() (*apicorev1.Secret, error) {

	return gsh.findSecret(context.TODO(), client.ObjectKey{
		Name:      gatewaySecretName,
		Namespace: istioNamespace,
	})
}

func (gsh *gatewaySecretHandler) newGatewaySecret(caBundle *apicorev1.Secret) *apicorev1.Secret {
	gwSecret := &apicorev1.Secret{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: apicorev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name:      gatewaySecretName,
			Namespace: istioNamespace,
		},
		Data: map[string][]byte{
			"tls.crt": caBundle.Data["root.tls.crt"],
			"tls.key": caBundle.Data["root.tls.key"],
			"ca.crt":  joinCACertsFromBundle(caBundle),
		},
	}
	return gwSecret
}

// joinCACertsFromBundle joins the CA certs from keys in the caBundle named "ca-bundle-0", "ca-bundle-1", etc. There are at most 3 keys.
func joinCACertsFromBundle(caBundle *apicorev1.Secret) []byte {
	caCerts := []byte{}

	if cert, ok := caBundle.Data["ca-bundle-0"]; ok {
		caCerts = append(caCerts, cert...)
	}

	if cert, ok := caBundle.Data["ca-bundle-1"]; ok {
		caCerts = append(caCerts, cert...)
	}

	//TODO: Not really needed, or is it?
	if cert, ok := caBundle.Data["ca-bundle-2"]; ok {
		caCerts = append(caCerts, cert...)
	}

	return caCerts
}
