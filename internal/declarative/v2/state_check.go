package v2

import (
	"context"
	"errors"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/lifecycle-manager/api/shared"
)

var ErrNotValidClientObject = errors.New("object in resource info is not a valid client object")

type StateCheck interface {
	GetState(ctx context.Context, clnt client.Client, resources []client.Object) (shared.State, error)
}

type ExistsStateCheck struct{}

func NewExistsStateCheck() *ExistsStateCheck {
	return &ExistsStateCheck{}
}

func (c *ExistsStateCheck) GetState(
	ctx context.Context,
	clnt client.Client,
	resources []client.Object,
) (shared.State, error) {
	for _, obj := range resources {
		if err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj); client.IgnoreNotFound(err) != nil {
			return shared.StateError, fmt.Errorf("failed to fetch object by key: %w", err)
		}
	}
	return shared.StateReady, nil
}
