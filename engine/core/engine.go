package core

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/grab/async/async"
	"golang.org/x/sync/errgroup"
)

var (
	planType = reflect.TypeOf((*plan)(nil)).Elem()
	preHookType = reflect.TypeOf((*pre)(nil)).Elem()
	postHookType = reflect.TypeOf((*post)(nil)).Elem()
)

type analyzedPlan struct {
	isSequential bool
	componentIDs []string
	preHooks []pre
	postHooks []post
}

type Engine struct {
	computers     map[string]computer
	syncComputers map[string]syncComputer
	plans         map[string]analyzedPlan
}

func NewEngine() Engine {
	return Engine{
		computers:     make(map[string]computer),
		syncComputers: make(map[string]syncComputer),
		plans:         make(map[string]analyzedPlan),
	}
}

func (e Engine) RegisterComputer(v any, c computer) {
	e.computers[extractFullNameFromValue(v)] = c
}

func (e Engine) RegisterSyncComputer(v any, c syncComputer) {
	e.syncComputers[extractFullNameFromValue(v)] = c
}

func (e Engine) IsRegistered(v any) bool {
	fullName := extractFullNameFromValue(v)

	_, ok := e.computers[fullName]
	if ok {
		return true
	}

	_, ok = e.syncComputers[fullName]
	return ok
}

func (e Engine) AnalyzePlan(p plan) string {
	val := reflect.ValueOf(p)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	var preHooks []pre
	var postHooks []post

	componentIDs := make([]string, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i).Type

		typeAndPointerTypeIsNotPlanType := !fieldType.Implements(planType) && !reflect.New(fieldType).Type().Implements(planType)

		if typeAndPointerTypeIsNotPlanType && fieldType.Implements(preHookType) {
			preHook := reflect.New(fieldType).Interface().(pre)
			preHooks = append(preHooks, preHook)

			continue
		}

		if typeAndPointerTypeIsNotPlanType && fieldType.Implements(postHookType) {
			postHook := reflect.New(fieldType).Interface().(post)
			postHooks = append(postHooks, postHook)

			continue
		}

		componentID := extractFullNameFromType(fieldType)
		componentIDs[i] = componentID
	}

	planName := extractFullNameFromValue(p)

	toUpdate := e.findExistingPlanOrCreate(planName)
	toUpdate.isSequential = p.IsSequential()
	toUpdate.componentIDs = componentIDs
	toUpdate.preHooks = append(toUpdate.preHooks, preHooks...)
	toUpdate.postHooks = append(toUpdate.postHooks, postHooks...)

	e.plans[planName] = toUpdate

	return planName
}

func (e Engine) IsAnalyzed(p plan) bool {
	_, ok := e.plans[extractFullNameFromValue(p)]
	return ok
}

func (e Engine) IsExecutable(masterPlan plan) (err error) {
	var verifyFn func(planName string)
	verifyFn = func(planName string) {
		ap := e.findAnalyzedPlan(planName)

		for _, componentID := range ap.componentIDs {
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("plan is not executable, %v", r)
						fmt.Println(string(debug.Stack()))
					}
				}()

				// If plan is not executable, 1 of the computer will panic
				if c, ok := e.computers[componentID]; ok {
					c.Compute(masterPlan)
				}

				if c, ok := e.syncComputers[componentID]; ok {
					c.Compute(context.Background(), masterPlan)
				}

				if _, ok := e.plans[componentID]; ok {
					verifyFn(componentID)
				}
			}()
		}
	}

	verifyFn(extractFullNameFromValue(masterPlan))


	return
}

func (e Engine) ConnectPreHook(p plan, hooks ...pre) {
	planName := extractFullNameFromValue(p)

	toUpdate := e.findExistingPlanOrCreate(planName)
	toUpdate.preHooks = append(toUpdate.preHooks, hooks...)

	e.plans[planName] = toUpdate
}

func (e Engine) ConnectPostHook(p plan, hooks ...post) {
	planName := extractFullNameFromValue(p)

	toUpdate := e.findExistingPlanOrCreate(planName)
	toUpdate.postHooks = append(toUpdate.postHooks, hooks...)

	e.plans[planName] = toUpdate
}

func (e Engine) Execute(ctx context.Context, planName string, p plan) error {
	if err := e.doExecute(ctx, planName, p, p.IsSequential()); err != nil {
		return swallowErrPlanExecutionEndingEarly(err)
	}

	return nil
}

func (e Engine) doExecute(ctx context.Context, planName string, p plan, isSequential bool) error {
	ap := e.findAnalyzedPlan(planName)

	for _, h := range ap.preHooks {
		if err := h.PreExecute(p); err != nil {
			return err
		}
	}

	err := func() error {
		if isSequential {
			return e.doExecuteSync(ctx, p, ap.componentIDs)
		}

		return e.doExecuteAsync(ctx, p, ap.componentIDs)
	}()

	if err != nil {
		return err
	}

	for _, h := range ap.postHooks {
		if err := h.PostExecute(p); err != nil {
			return err
		}
	}

	return nil
}

func (e Engine) doExecuteSync(ctx context.Context, p plan, componentIDs []string) error {
	for _, componentID := range componentIDs {
		c, ok := e.syncComputers[componentID]
		if ok {
			if err := c.Compute(ctx, p); err != nil {
				return err
			}

			continue
		}

		// Nested plan gets executed synchronously
		if ap, ok := e.plans[componentID]; ok {
			if err := e.doExecute(ctx, componentID, p, ap.isSequential); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e Engine) doExecuteAsync(ctx context.Context, p plan, componentIDs []string) error {
	tasks := make([]async.SilentTask, 0, len(componentIDs))
	for _, componentID := range componentIDs {
		if c, ok := e.computers[componentID]; ok {
			// Compute() will create and assign all tasks into plan so that
			// when we call Execute() on any tasks, we won't get nil panic
			// due to task fields not yet initialized.
			tasks = append(tasks, c.Compute(p))

			continue
		}

		// Nested plan gets executed asynchronously by wrapping it inside a task
		if ap, ok := e.plans[componentID]; ok {
			task := async.NewSilentTask(
				func(ctx context.Context) error {
					return e.doExecute(ctx, componentID, p, ap.isSequential)
				},
			)

			tasks = append(tasks, task)
		}
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

func (e Engine) findExistingPlanOrCreate(planName string) analyzedPlan {
	if existing, ok := e.plans[planName]; ok {
		return existing
	}

	return analyzedPlan{}
}

func (e Engine) findAnalyzedPlan(planName string) analyzedPlan {
	ap, ok := e.plans[planName]
	if !ok || len(ap.componentIDs) == 0 {
		panic(ErrAnalyzePlanNotDone)
	}

	return ap
}
