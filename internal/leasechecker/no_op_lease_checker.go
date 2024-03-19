package leasechecker

import (
	"github.com/intuit/naavik/internal/types/context"
)

/*
Default implementation of the interface defined for DR.
*/
type noOpStateChecker struct{}

func (noOpStateChecker) ShouldRunOnIndependentGoRoutine() bool {
	return false
}

func (noOpStateChecker) InitStateCache(_ interface{}) error {
	return nil
}

func (noOpStateChecker) RunStateCheck(_ context.Context) {
	currentLeaseState.ReadOnly = readWriteEnabled
	currentLeaseState.IsStateInitialized = stateInitialized
}
