package skrresources

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
)

type ResourceInfoConverter interface {
	ResourceInfo(obj *unstructured.Unstructured) (*resource.Info, error)
}
