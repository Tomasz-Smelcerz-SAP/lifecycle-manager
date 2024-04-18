package zerodw

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	apicorev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	caBundleSecretName = "ca-bundle"
	kcpNamespace       = "kcp-system"
	kcpRootSecretName  = "klm-watcher-root-secret"
	istioNamespace     = "istio-system"
)

var (
	RootSecretNotFound = errors.New("Root secret not found")
)

type caBundleHandler struct {
	*secretManager
}

func NewCABundleHandler(kcpClient client.Client, log logr.Logger) *caBundleHandler {
	return &caBundleHandler{
		secretManager: &secretManager{
			kcpClient: kcpClient,
			log:       log,
		},
	}
}

func (cab *caBundleHandler) ManageCABundleSecret() error {

	caBundle, err := cab.FindCABundleSecret()

	if isNotFound(err) {
		// ca-bundle secret does not exist
		return cab.handleNonExisting()
	}
	if err != nil {
		return err
	}

	// ca-bundle secret exists
	return cab.handleExisting(caBundle)
}

func (cab *caBundleHandler) handleNonExisting() error {
	rootSecret, err := cab.findRootSecret()
	if isNotFound(err) {
		return RootSecretNotFound
	}

	if err != nil {
		return err
	}

	caBundle := cab.newEmptyCABundleSecret()

	caBundle.Data["root.tls.crt"] = rootSecret.Data["tls.crt"]
	caBundle.Data["root.tls.key"] = rootSecret.Data["tls.key"]
	caBundle.Data["root.ca.crt"] = rootSecret.Data["ca.crt"]

	//The "Data" field can't keep an array of values, so an integer key suffix is used to represent the array entries
	caBundle.Data["ca-bundle-0"] = rootSecret.Data["ca.crt"]
	//TODO: Just for testing, remove tls.crt
	caBundle.Data["ca-bundle-1"] = rootSecret.Data["tls.crt"]
	return cab.kcpClient.Create(context.TODO(), caBundle)
}

func (cab *caBundleHandler) handleExisting(caBundle *apicorev1.Secret) error {
	/*
		rootSecret, err := cab.findRootSecret()
		if isNotFound(err) {
			return RootSecretNotFound
		}

		if err != nil {
			return err
		}
	*/
	return nil
}

// func newGatewaySecret creates a gateway Secret
func (cab *caBundleHandler) newGatewaySecret() *apicorev1.Secret {
	return &apicorev1.Secret{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: apicorev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name:      "gateway-secret",
			Namespace: istioNamespace,
		},
		Type: apicorev1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
}

// newEmptyCABundleSecret creates a CA bundle secret
func (cab *caBundleHandler) newEmptyCABundleSecret() *apicorev1.Secret {
	return &apicorev1.Secret{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: apicorev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name:      caBundleSecretName,
			Namespace: kcpNamespace,
		},
		Type: apicorev1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
}

// FindCABundleSecret finds the CA bundle secret
func (cab *caBundleHandler) FindCABundleSecret() (*apicorev1.Secret, error) {
	return cab.findSecret(context.TODO(), client.ObjectKey{
		Name:      caBundleSecretName,
		Namespace: kcpNamespace,
	})
}

func (cab *caBundleHandler) findRootSecret() (*apicorev1.Secret, error) {
	return cab.findSecret(context.TODO(), client.ObjectKey{
		Name:      kcpRootSecretName,
		Namespace: istioNamespace,
	})
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return client.IgnoreNotFound(err) == nil
}
