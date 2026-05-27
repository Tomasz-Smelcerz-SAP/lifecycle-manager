package v2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apicorev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	declarativev2 "github.com/kyma-project/lifecycle-manager/internal/declarative/v2"
)

func newObj(name, namespace, kind string) client.Object {
	obj := &apicorev1.Pod{
		ObjectMeta: apimetav1.ObjectMeta{Name: name, Namespace: namespace},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Kind: kind})
	return obj
}

func TestResourceList_Difference(t *testing.T) {
	dummyPod := newObj("foo", "default", "Pod")
	dummyService := newObj("bar", "default", "Service")
	dummyDeploy := newObj("baz", "default", "Deployment")

	list1 := declarativev2.ResourceList{dummyPod, dummyService, dummyDeploy}
	list2 := declarativev2.ResourceList{dummyService}

	diff := list1.Difference(list2)

	assert.Len(t, diff, 2)
	assert.Contains(t, diff, dummyPod)
	assert.Contains(t, diff, dummyDeploy)
	assert.NotContains(t, diff, dummyService)
}
