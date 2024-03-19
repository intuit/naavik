package leasechecker

import (
	"sync"

	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
)

const (
	readWriteEnabled    = false
	readOnlyEnabled     = true
	stateNotInitialized = false
	stateInitialized    = true
)

type LeaseState struct {
	ReadOnly           bool
	IsStateInitialized bool
}

var (
	leaseLock         sync.Mutex
	currentLeaseState = &LeaseState{
		ReadOnly: readOnlyEnabled,
	}
)

type LeaseStateChecker interface {
	RunStateCheck(ctx context.Context)
	ShouldRunOnIndependentGoRoutine() bool
	InitStateCache(interface{}) error
}

/*
Utility function to start DR checks.
DR checks can be run either on the main go routine or a new go routine.
*/
func RunStateCheck(ctx context.Context, leaseStateChecker LeaseStateChecker) {
	ctx.Log.Info("Starting state checker")
	if leaseStateChecker.ShouldRunOnIndependentGoRoutine() {
		ctx.Log.Info("Starting state checker on a Go Routine")
		go leaseStateChecker.RunStateCheck(ctx)
	} else {
		ctx.Log.Info("Starting state checker on existing Go Routine")
		leaseStateChecker.RunStateCheck(ctx)
	}
}

func GetStateChecker(ctx context.Context, stateChecker string) LeaseStateChecker {
	switch stateChecker {
	case types.StateCheckerNone:
		ctx.Log.Str(logger.NameKey, stateChecker).Info("Initializing NoOp based state checker")
		return noOpStateChecker{}
	default:
		ctx.Log.Fatalf("invalid state checker %q", stateChecker)
	}

	return nil
}

func IsReadOnly() bool {
	leaseLock.Lock()
	defer leaseLock.Unlock()
	return currentLeaseState.ReadOnly
}

func IsStateInitialized() bool {
	leaseLock.Lock()
	defer leaseLock.Unlock()
	return currentLeaseState.IsStateInitialized
}

// ResetState resets the state of the lease checker. This is used for testing purposes only.
func ResetState() {
	leaseLock.Lock()
	defer leaseLock.Unlock()
	currentLeaseState = &LeaseState{
		ReadOnly: readOnlyEnabled,
	}
}
