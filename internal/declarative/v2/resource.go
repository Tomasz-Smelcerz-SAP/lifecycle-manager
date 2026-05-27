package v2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceList provides convenience methods for comparing collections of client objects.
type ResourceList []client.Object

// Difference will return a new ResourceList with objects not contained in rs.
func (r ResourceList) Difference(rs ResourceList) ResourceList {
	return r.filter(func(obj client.Object) bool {
		return !rs.contains(obj)
	})
}

// append adds an object to the ResourceList.
func (r *ResourceList) append(val client.Object) {
	*r = append(*r, val)
}

// filter returns a new ResourceList with objects that satisfy the predicate fn.
func (r ResourceList) filter(fn func(client.Object) bool) ResourceList {
	var result ResourceList
	for _, obj := range r {
		if fn(obj) {
			result.append(obj)
		}
	}
	return result
}

// contains checks to see if an object exists.
func (r ResourceList) contains(obj client.Object) bool {
	for _, i := range r {
		if isMatchingObject(i, obj) {
			return true
		}
	}
	return false
}

// isMatchingObject returns true if objects match on Name, Namespace and Kind.
func isMatchingObject(a, b client.Object) bool {
	if a == nil || b == nil {
		return false
	}
	return a.GetName() == b.GetName() &&
		a.GetNamespace() == b.GetNamespace() &&
		getKind(a.GetObjectKind()) == getKind(b.GetObjectKind())
}

func getKind(obknd schema.ObjectKind) string {
	if obknd == nil {
		return ""
	}
	return obknd.GroupVersionKind().Kind
}
