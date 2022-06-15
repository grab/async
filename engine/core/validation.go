package core

import (
	"context"
	"fmt"
	"runtime/debug"
)

func (e Engine) IsAnalyzed(p plan) bool {
	_, ok := e.plans[extractFullNameFromValue(p)]
	return ok
}

func (e Engine) IsRegistered(v any) bool {
	fullName := extractFullNameFromValue(v)

	_, ok := e.computers[fullName]
	return ok
}

func (e Engine) IsExecutable(p masterPlan) (err error) {
	var verifyFn func(planName string)
	verifyFn = func(planName string) {
		ap := e.findAnalyzedPlan(planName)

		for _, component := range ap.components {
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("plan is not executable, %v", r)
						fmt.Println(string(debug.Stack()))
					}
				}()

				// If plan is not executable, 1 of the computer will panic
				if c, ok := e.computers[component.id]; ok {
					c.Compute(context.Background(), p)
				}

				if _, ok := e.plans[component.id]; ok {
					verifyFn(component.id)
				}
			}()
		}
	}

	verifyFn(extractFullNameFromValue(p))

	return
}
