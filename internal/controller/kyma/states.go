package kyma

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/kyma-project/lifecycle-manager/pkg/log"
	"github.com/kyma-project/lifecycle-manager/pkg/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

// KymaStateMachine is a state machine for Kyma.
type KymaStateMachine interface {
	WithLogger(logger logr.Logger) KymaStateMachine
	Run(ctx context.Context, kyma *v1beta2.Kyma) (ctrl.Result, error)
}

type StateMachineImpl struct {
	ctx          context.Context
	logger       logr.Logger
	req          ctrl.Request
	initialState State
}

func (s *StateMachineImpl) WithLogger(logger logr.Logger) *StateMachineImpl {
	s.logger = logger
	return s
}

func (s *StateMachineImpl) Run(ctx context.Context, kyma *v1beta2.Kyma) (ctrl.Result, error) {
	currentState := s.getInitialState()
	for {
		res, err := currentState.Run(ctx, kyma)
		if err != nil {
			//TODO: log error
			return ctrl.Result{}, err
		}
		if stateResult.QuitNow() {
			return stateResult.result, stateResult.err
		}
		currentState = s.getNextState(currentState.Name(), res)
	}
}

func (s *StateMachineImpl) getInitialState() State {
	return s.initialState
}

func (s *StateMachineImpl) getNextState(currentState string) string {
	switch currentState {
	case "InitKyma":
		if s.SyncKymaEnabled() {
			return "InitSkrSync"
		}
		return "CheckForDeletion"
	case "InitSkrSync":
		return "CheckForDeletion"
	case "CheckForDeletion":
		return "EnsureMetadata"
	case "EnsureMetadata":
		if s.SyncKymaEnabled() {
			return "SyncRemoteCRDs"
		}
		return "ProcessKymaState"
	case "ProcessKymaState":
		if s.SyncKymaEnabled() {
			return "SyncStatusToRemote"
		}
		return "[Quit]"
	default:
		"InitKyma"
	}

}

type StateResult struct {
	result ctrl.Result
	err    error
}

func SuccessResult() *StateResult {
	return &StateResult{result: ctrl.Result{RequeueAfter: r.RequeueIntervals.Success}, err: nil}
}

type State interface {
	Run(ctx context.Context, logger logr.Logger) (ctrl.Result, error)
}

type InitKyma struct {
	logger         logr.Logger
	kymaService    KymaService
	watcherService WatcherService
}

func (st *InitKyma) Run(ctx context.Context, kyma *v1beta2.Kyma) (*StateResult, error) {
	status.InitConditions(kyma, st.kymaService.SyncKymaEnabled(kyma), st.watcherService.WatcherEnabled(kyma))

	if kyma.SkipReconciliation() {
		st.logger.V(log.DebugLevel).Info("skipping reconciliation for Kyma: " + kyma.Name)
		return SuccessResult(), nil
	}
}
