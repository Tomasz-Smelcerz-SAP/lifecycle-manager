package zerodw

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	apicorev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	lastModifiedAtAnnotation = "lastModifiedAt"
)

type secretManager struct {
	log       logr.Logger
	kcpClient client.Client
}

func (sm *secretManager) updateLastModifiedAt(secret *apicorev1.Secret) {
	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}
	secret.Annotations[lastModifiedAtAnnotation] = apimetav1.Now().Format(time.RFC3339)
}

func (sm *secretManager) findSecret(ctx context.Context, objKey client.ObjectKey) (*apicorev1.Secret, error) {
	secret := &apicorev1.Secret{}

	err := sm.kcpClient.Get(ctx, objKey, secret)

	if err != nil {
		return nil, err
	}

	return secret, nil
}
