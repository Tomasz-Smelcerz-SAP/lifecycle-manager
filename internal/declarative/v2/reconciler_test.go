//nolint:testpackage // test private functions
package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiappsv1 "k8s.io/api/apps/v1"
	apicorev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func objectWithKind(obj client.Object, kind string) client.Object {
	obj.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Kind: kind})
	return obj
}

func TestPruneResource(t *testing.T) {
	t.Parallel()
	kubeNs := objectWithKind(&apicorev1.Namespace{
		ObjectMeta: apimetav1.ObjectMeta{Name: "kube-system"},
	}, "Namespace")
	service := objectWithKind(&apicorev1.Service{
		ObjectMeta: apimetav1.ObjectMeta{Name: "some-service"},
	}, "Service")
	kymaNs := objectWithKind(&apicorev1.Namespace{
		ObjectMeta: apimetav1.ObjectMeta{Name: "kyma-system"},
	}, "Namespace")
	deployment := objectWithKind(&apiappsv1.Deployment{
		ObjectMeta: apimetav1.ObjectMeta{Name: "some-deploy"},
	}, "Deployment")
	crd := objectWithKind(&apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: apimetav1.ObjectMeta{Name: "btpoperator"},
	}, "CustomResourceDefinition")

	t.Run("contains kyma-system", func(t *testing.T) {
		t.Parallel()

		infos := []client.Object{kubeNs, service, kymaNs, deployment}

		result := pruneResource(infos, "Namespace", namespaceNotBeRemoved)

		require.Len(t, result, 3)
		require.NotContains(t, result, kymaNs)
	})

	t.Run("prune a crd", func(t *testing.T) {
		t.Parallel()

		infos := []client.Object{kubeNs, service, kymaNs, deployment, crd}

		result := pruneResource(infos, "CustomResourceDefinition", "btpoperator")

		require.Len(t, result, 4)
		require.NotContains(t, result, crd)
	})

	t.Run("does not contain kyma-system", func(t *testing.T) {
		t.Parallel()

		infos := []client.Object{kubeNs, service, deployment}

		result := pruneResource(infos, "Namespace", namespaceNotBeRemoved)

		require.Len(t, result, 3)
		require.Contains(t, result, kubeNs)
		require.Contains(t, result, service)
		require.Contains(t, result, deployment)
	})
}
