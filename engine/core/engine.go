package core

import (
	"context"
	"fmt"
	"reflect"

	"github.com/grab/async/async"
	"golang.org/x/sync/errgroup"
)

var (
	planType        = reflect.TypeOf((*plan)(nil)).Elem()
	preHookType     = reflect.TypeOf((*pre)(nil)).Elem()
	postHookType    = reflect.TypeOf((*post)(nil)).Elem()
	asyncResultType = reflect.TypeOf(AsyncResult{})
)

type parsedComponent struct {
	id     string
	setter reflect.Method
}

type analyzedPlan struct {
	isSequential bool
	components   []parsedComponent
	preHooks     []pre
	postHooks    []post
}

type Engine struct {
	computers map[string]computer
	plans     map[string]analyzedPlan
}

func NewEngine() Engine {
	return Engine{
		computers: make(map[string]computer),
		plans:     make(map[string]analyzedPlan),
	}
}

func (e Engine) RegisterComputer(v any, c computer) {
	e.computers[extractFullNameFromValue(v)] = c
}

func (e Engine) RegisterSilentComputer(v any, sc silentComputer) {
	e.computers[extractFullNameFromValue(v)] = bridgeComputer{
		sc: sc,
	}
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

func (e Engine) AnalyzePlan(p plan) string {
	val := reflect.ValueOf(p)
	if val.Kind() != reflect.Pointer {
		panic(ErrPlanMustUsePointerReceiver)
	}

	val = val.Elem()
	pType := reflect.ValueOf(p).Type()

	var preHooks []pre
	var postHooks []post

	components := make([]parsedComponent, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i).Type
		fieldPointerType := reflect.PointerTo(fieldType)

		// Hook types might be embedded in a parent plan struct. Hence, we need to check if the type
		// is a hook but not a plan so that we don't register a plan as a hook.
		typeAndPointerTypeIsNotPlanType := !fieldType.Implements(planType) && !reflect.New(fieldType).Type().Implements(planType)

		// Hooks might be implemented with value or pointer receivers.
		isPreHookType := fieldType.Implements(preHookType) || fieldPointerType.Implements(preHookType)
		isPostHookType := fieldType.Implements(postHookType) || fieldPointerType.Implements(postHookType)

		if typeAndPointerTypeIsNotPlanType && isPreHookType {
			preHook := reflect.New(fieldType).Interface().(pre)
			preHooks = append(preHooks, preHook)

			continue
		}

		if typeAndPointerTypeIsNotPlanType && isPostHookType {
			postHook := reflect.New(fieldType).Interface().(post)
			postHooks = append(postHooks, postHook)

			continue
		}

		componentID := extractFullNameFromType(fieldType)

		component := func() parsedComponent {
			if fieldType.ConvertibleTo(asyncResultType) {
				if p.IsSequential() {
					panic(fmt.Errorf("sequential plan cannot contain AsyncResult field: %s", extractShortName(componentID)))
				}

				if setter, ok := pType.MethodByName("Set" + extractShortName(componentID)); ok {
					return parsedComponent{
						id:     componentID,
						setter: setter,
					}
				}

				panic(fmt.Errorf("parallel plan must have setter for AsyncResult field: %s", extractShortName(componentID)))
			}

			return parsedComponent{
				id: componentID,
			}
		}()

		components[i] = component
	}

	planName := extractFullNameFromValue(p)

	toUpdate := e.findExistingPlanOrCreate(planName)
	toUpdate.isSequential = p.IsSequential()
	toUpdate.components = components
	toUpdate.preHooks = append(toUpdate.preHooks, preHooks...)
	toUpdate.postHooks = append(toUpdate.postHooks, postHooks...)

	e.plans[planName] = toUpdate

	return planName
}

func (e Engine) Execute(ctx context.Context, planName string, p masterPlan) error {
	if err := e.doExecute(ctx, planName, p, p.IsSequential()); err != nil {
		return swallowErrPlanExecutionEndingEarly(err)
	}

	return nil
}

func (e Engine) doExecute(ctx context.Context, planName string, p masterPlan, isSequential bool) error {
	ap := e.findAnalyzedPlan(planName)

	for _, h := range ap.preHooks {
		if err := h.PreExecute(p); err != nil {
			return err
		}
	}

	err := func() error {
		if isSequential {
			return e.doExecuteSync(ctx, p, ap.components)
		}

		return e.doExecuteAsync(ctx, p, ap.components)
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

func (e Engine) doExecuteSync(ctx context.Context, p masterPlan, components []parsedComponent) error {
	for _, component := range components {
		if c, ok := e.computers[component.id]; ok {
			task := async.NewTask(
				func(taskCtx context.Context) (any, error) {
					return c.Compute(taskCtx, p)
				},
			)

			if err := task.ExecuteSync(ctx).Error(); err != nil {
				return err
			}

			continue
		}

		// Nested plan gets executed synchronously
		if ap, ok := e.plans[component.id]; ok {
			if err := e.doExecute(ctx, component.id, p, ap.isSequential); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e Engine) doExecuteAsync(ctx context.Context, p masterPlan, components []parsedComponent) error {
	tasks := make([]async.SilentTask, 0, len(components))
	for _, component := range components {
		componentID := component.id

		if c, ok := e.computers[componentID]; ok {
			task := async.NewTask(
				func(taskCtx context.Context) (any, error) {
					return c.Compute(taskCtx, p)
				},
			)

			tasks = append(tasks, task)

			// Register AsyncResult in a parallel plan's field
			component.setter.Func.Call([]reflect.Value{reflect.ValueOf(p), reflect.ValueOf(newAsyncResult(task))})

			continue
		}

		// Nested plan gets executed asynchronously by wrapping it inside a task
		if ap, ok := e.plans[componentID]; ok {
			task := async.NewSilentTask(
				func(taskCtx context.Context) error {
					return e.doExecute(taskCtx, componentID, p, ap.isSequential)
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
				return t.ExecuteSync(groupCtx).Error()
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
	if !ok || len(ap.components) == 0 {
		panic(ErrPlanNotAnalyzed)
	}

	return ap
}
