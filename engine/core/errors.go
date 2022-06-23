package core

import "errors"

var (
	// ErrPlanExecutionEndingEarly can be thrown actively by clients to end plan execution early.
	// For example, a value was retrieved from cache and thus, there's no point executing the algo
	// to calculate this value anymore. The engine will swallow this error, end execution and then
	// return a nil error to clients.
	ErrPlanExecutionEndingEarly   = errors.New("plan ending early")
	ErrPlanNotAnalyzed            = errors.New("plan must be analyzed before getting executed")
	ErrPlanMustUsePointerReceiver = errors.New("the passed in plan must be a pointer")
)
