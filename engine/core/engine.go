package core

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/grab/async/async"
	"golang.org/x/sync/errgroup"
)

type Engine struct {
	computers     map[string]computer
	syncComputers map[string]syncComputer
	plans         map[string][]string
}

func NewEngine() Engine {
	return Engine{
		computers:     make(map[string]computer),
		syncComputers: make(map[string]syncComputer),
		plans:         make(map[string][]string),
	}
}

func (e Engine) RegisterComputer(v any, c computer) {
	e.computers[e.extractFullNameFromValue(v)] = c
}

func (e Engine) RegisterSyncComputer(v any, c syncComputer) {
	e.syncComputers[e.extractFullNameFromValue(v)] = c
}

func (e Engine) IsRegistered(v any) bool {
	fullName := e.extractFullNameFromValue(v)

	_, ok := e.computers[fullName]
	if ok {
		return true
	}

	_, ok = e.syncComputers[fullName]
	return ok
}

func (e Engine) extractFullNameFromValue(v any) string {
	if reflect.ValueOf(v).Kind() == reflect.Pointer {
		t := reflect.ValueOf(v).Elem().Type()
		return e.extractFullNameFromType(t)
	}

	t := reflect.TypeOf(v)
	return e.extractFullNameFromType(t)
}

func (e Engine) extractFullNameFromType(t reflect.Type) string {
	return t.PkgPath() + "/" + t.Name()
}

func (e Engine) IsAnalyzed(p plan) bool {
	_, ok := e.plans[e.extractFullNameFromValue(p)]
	return ok
}

func (e Engine) AnalyzePlan(p plan) string {
	val := reflect.ValueOf(p)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	computers := make([]string, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		computerID := e.extractFullNameFromType(val.Type().Field(i).Type)
		computers[i] = computerID
	}

	planName := e.extractFullNameFromValue(p)
	e.plans[planName] = computers

	return planName
}

func (e Engine) IsPlanExecutable(p plan) (err error) {
	planName := e.extractFullNameFromValue(p)

	computers, ok := e.plans[planName]
	if !ok {
		panic(ErrAnalyzePlanNotDone)
	}

	for _, computerID := range computers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("plan is not executable, %v", r)
					fmt.Println(string(debug.Stack()))
				}
			}()

			// If plan is not executable, 1 of the computer will panic
			c, ok := e.computers[computerID]
			if ok {
				c.Compute(p)
			}
		}()
	}

	return
}

func (e Engine) Execute(ctx context.Context, masterPlan any, planName string, p plan) error {
	if pre, ok := p.(prePlan); ok {
		if err := pre.PreExecute(ctx, masterPlan); err != nil {
			return e.swallowErrPlanExecutionEndingEarly(err)
		}
	}

	if err := e.doExecute(ctx, planName, p); err != nil {
		return e.swallowErrPlanExecutionEndingEarly(err)
	}

	if post, ok := p.(postPlan); ok {
		if err := post.PostExecute(ctx, masterPlan); err != nil {
			return e.swallowErrPlanExecutionEndingEarly(err)
		}
	}

	return nil
}

func (e Engine) swallowErrPlanExecutionEndingEarly(err error) error {
	// Execution was intentionally ended by clients
	if err == ErrPlanExecutionEndingEarly {
		return nil
	}

	return err
}

func (e Engine) doExecute(ctx context.Context, planName string, p plan) error {
	computers, ok := e.plans[planName]
	if !ok {
		panic(ErrAnalyzePlanNotDone)
	}

	if p.IsSequential() {
		if err := e.doExecuteSync(ctx, p, computers); err != nil {
			return err
		}

		return nil
	}

	if err := e.doExecuteAsync(ctx, p, computers); err != nil {
		return err
	}

	return nil
}

func (e Engine) doExecuteSync(ctx context.Context, p plan, computers []string) error {
	for _, computerID := range computers {
		c, ok := e.syncComputers[computerID]
		if !ok {
			continue
		}

		if err := c.Compute(ctx, p); err != nil {
			return err
		}
	}

	return nil
}

func (e Engine) doExecuteAsync(ctx context.Context, p plan, computers []string) error {
	tasks := make([]async.SilentTask, 0, len(computers))
	for _, computerID := range computers {
		c, ok := e.computers[computerID]
		if !ok {
			continue
		}

		// Compute() will create and assign all tasks into plan so that
		// when we call Execute() on any tasks, we won't get nil panic
		// due to task fields not yet initialized.
		tasks = append(tasks, c.Compute(p))
	}

	g, groupCtx := errgroup.WithContext(ctx)
	for _, task := range tasks {
		t := task
		g.Go(
			func() error {
				t.ExecuteSync(groupCtx)

				return t.Error()
			},
		)
	}

	return g.Wait()
}
