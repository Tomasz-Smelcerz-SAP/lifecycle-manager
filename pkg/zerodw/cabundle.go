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
	CaBundleSecretName = "ca-bundle"
	KcpNamespace       = "kcp-system"
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
	rootSecret, err := cab.findKcpRootSecret()
	if isNotFound(err) {
		// root secret not found. Wait until it is created
		cab.log.Error(RootSecretNotFound, "caBundleHandler")
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
	//caBundle.Data["ca-bundle-1"] = rootSecret.Data["tls.crt"]

	err = cab.create(context.TODO(), caBundle)
	if err == nil {
		cab.log.Info("created the caBundle secret", "reason", "caBundle secret does not exist")
	}
	return err
}

func (cab *caBundleHandler) handleExisting(caBundle *apicorev1.Secret) error {
	rootSecret, err := cab.findKcpRootSecret()
	if isNotFound(err) {
		cab.log.Error(RootSecretNotFound, "caBundleHandler")
		// root secret not found. Wait until it is created
		return RootSecretNotFound
	}
	if err != nil {
		return err
	}

	//Is the migration already done?
	//TODO: Replace with a proper mechanism once it's determined how to reliably tell if all the clients are migrated...
	allClientsMigrated, ok := caBundle.Annotations[AllClientsMigrated]
	if ok && allClientsMigrated == "true" {
		cab.log.Info("Finishing migration", "reason", "all clients are migrated")
		delete(caBundle.Data, "ca-bundle-0") //it should be the same as rootSecret.Data["ca.crt"]. We could leave it but it's more safer to explicitly re-assign
		delete(caBundle.Data, "ca-bundle-1")
		delete(caBundle.Annotations, MigrationPendingAnnotation)
		delete(caBundle.Annotations, AllClientsMigrated)

		caBundle.Data["root.tls.crt"] = rootSecret.Data["tls.crt"]
		caBundle.Data["root.tls.key"] = rootSecret.Data["tls.key"]
		caBundle.Data["root.ca.crt"] = rootSecret.Data["ca.crt"]
		caBundle.Data["ca-bundle-0"] = rootSecret.Data["ca.crt"]

		return cab.update(context.TODO(), caBundle)
	}

	lastModifiedAtValue, ok := caBundle.Annotations[LastModifiedAtAnnotation]
	if ok {
		caBundleSecretLastModifiedAt, err := time.Parse(time.RFC3339, lastModifiedAtValue)
		if err != nil {
			return err
		}
		//TODO: Is CreationTimestamp change enough to detect a secret rotation?
		//TODO: Maybe we should in addition compare the secret data to detect if the certificate keys really changed
		if rootSecret.CreationTimestamp.Time.After(caBundleSecretLastModifiedAt) {
			cab.log.Info("Starting migration", "reason", "Root secret is more recent than the CA bundle secret")
			caBundle.Data["ca-bundle-1"] = caBundle.Data["ca-bundle-0"]
			caBundle.Data["ca-bundle-0"] = rootSecret.Data["ca.crt"]
			caBundle.Annotations[MigrationPendingAnnotation] = "true"
			return cab.update(context.TODO(), caBundle)
		}
	}

	return nil
}

// newEmptyCABundleSecret creates a CA bundle secret
func (cab *caBundleHandler) newEmptyCABundleSecret() *apicorev1.Secret {
	return &apicorev1.Secret{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: apicorev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name:      CaBundleSecretName,
			Namespace: KcpNamespace,
		},
		Type: apicorev1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
}

// FindCABundleSecret finds the CA bundle secret
func (cab *caBundleHandler) FindCABundleSecret() (*apicorev1.Secret, error) {
	return cab.findSecret(context.TODO(), client.ObjectKey{
		Name:      CaBundleSecretName,
		Namespace: KcpNamespace,
	})
}

// findKcpRootSecret finds the KCP root secret
func (cab *caBundleHandler) findKcpRootSecret() (*apicorev1.Secret, error) {
	return cab.findSecret(context.TODO(), client.ObjectKey{
		Name:      kcpRootSecretName,
		Namespace: istioNamespace,
	})
}
