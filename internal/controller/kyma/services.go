package kyma

import "github.com/kyma-project/lifecycle-manager/api/v1beta2"

type KymaService interface {
	SyncKymaEnabled(kyma *v1beta2.Kyma) bool
}

type WatcherService interface {
	WatcherEnabled(kyma *v1beta2.Kyma) bool
}
