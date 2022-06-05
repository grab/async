package core

import "errors"

var (
	ErrInvalidPlan        = errors.New("computer cannot handle the given plan, input/output is missing")
	ErrAnalyzePlanNotDone = errors.New("plan must be analyzed before getting executed")
)
