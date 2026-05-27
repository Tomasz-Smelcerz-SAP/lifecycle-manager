package resources

import (
	"context"
	"errors"

	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/kyma-project/lifecycle-manager/pkg/util"
)

var ErrDeletionNotFinished = errors.New("deletion is not yet finished")

type ConcurrentCleanup struct {
	clnt     client.Client
	manifest *v1beta2.Manifest
}

func NewConcurrentCleanup(clnt client.Client, manifest *v1beta2.Manifest) *ConcurrentCleanup {
	return &ConcurrentCleanup{
		clnt:     clnt,
		manifest: manifest,
	}
}

func (c *ConcurrentCleanup) DeleteDiffResources(ctx context.Context, resources []client.Object) error {
	status := c.manifest.GetStatus()
	operatorRelatedResources, operatorManagedResources := SplitResources(resources)
	if err := c.cleanupResources(ctx, operatorManagedResources, status); err != nil {
		return err
	}
	return c.cleanupResources(ctx, operatorRelatedResources, status)
}

func SplitResources(resources []client.Object) ([]client.Object, []client.Object) {
	operatorRelatedResources := make([]client.Object, 0)
	operatorManagedResources := make([]client.Object, 0)

	for _, obj := range resources {
		if IsOperatorRelatedResources(obj.GetObjectKind().GroupVersionKind().Kind) {
			operatorRelatedResources = append(operatorRelatedResources, obj)
			continue
		}
		operatorManagedResources = append(operatorManagedResources, obj)
	}

	return operatorRelatedResources, operatorManagedResources
}

func IsOperatorRelatedResources(kind string) bool {
	return kind == "CustomResourceDefinition" ||
		kind == "Namespace" ||
		kind == "ServiceAccount" ||
		kind == "Role" ||
		kind == "ClusterRole" ||
		kind == "RoleBinding" ||
		kind == "ClusterRoleBinding" ||
		kind == "Service" ||
		kind == "Deployment"
}

func (c *ConcurrentCleanup) Run(ctx context.Context, objs []client.Object) error {
	objsCount := len(objs)
	results := make(chan error, objsCount)
	for i := range objs {
		go c.cleanupResource(ctx, objs[i], results)
	}

	var errs []error
	for range objs {
		err := <-results
		if util.IsNotFound(err) {
			objsCount--
			continue
		}
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if objsCount > 0 {
		return ErrDeletionNotFinished
	}
	return nil
}

func (c *ConcurrentCleanup) cleanupResources(
	ctx context.Context,
	resources []client.Object,
	status shared.Status,
) error {
	if err := c.Run(ctx, resources); errors.Is(err, ErrDeletionNotFinished) {
		c.manifest.SetStatus(status.WithState(shared.StateWarning).WithErr(err))
		return err
	} else if err != nil {
		c.manifest.SetStatus(status.WithState(shared.StateError).WithErr(err))
		return err
	}
	return nil
}

func (c *ConcurrentCleanup) cleanupResource(ctx context.Context, obj client.Object, results chan error) {
	results <- c.clnt.Delete(ctx, obj, client.PropagationPolicy(apimetav1.DeletePropagationBackground))
}
