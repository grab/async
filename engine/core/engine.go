package core

import (
    "context"
    "fmt"
    "reflect"
    "runtime/debug"

    "golang.org/x/sync/errgroup"
)

type Engine struct {
    computers map[string]computer
    plans map[string][]string
}

func NewEngine() Engine {
    return Engine{
        computers: make(map[string]computer),
        plans: make(map[string][]string),
    }
}

func (e Engine) RegisterComputer(v any, c computer) {
    e.computers[e.extractFullNameFromValue(v)] = c
}

func (e Engine) IsRegistered(v any) bool {
    _, ok := e.computers[e.extractFullNameFromValue(v)]
    return ok
}

func (e Engine) extractFullNameFromValue(v any) string {
    if reflect.ValueOf(v).Kind() == reflect.Pointer {
        t := reflect.ValueOf(v).Elem().Type()
        return t.PkgPath()+"/"+t.Name()
    }

    t := reflect.TypeOf(v)
    return t.PkgPath()+"/"+t.Name()
}

func (e Engine) extractFullNameFromType(t reflect.Type) string {
    return t.PkgPath()+"/"+t.Name()
}

func (e Engine) IsAnalyzed(plan any) bool {
    _, ok := e.plans[e.extractFullNameFromValue(plan)]
    return ok
}

func (e Engine) AnalyzePlan(plan any) string {
    var computers []string

    val := reflect.ValueOf(plan)
    for i := 0 ; i < val.NumField() ; i++ {
        computerOutputName := e.extractFullNameFromType(val.Type().Field(i).Type)

        computers = append(computers, computerOutputName)
    }

    planName := e.extractFullNameFromValue(plan)
    e.plans[planName] = computers

    return planName
}

func (e Engine) IsPlanExecutable(plan any) (err error) {
    planName := e.extractFullNameFromValue(plan)

    computers, ok := e.plans[planName]
    if !ok {
        panic(ErrAnalyzePlanNotDone)
    }

    for _, computerOutputName := range computers {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    err = fmt.Errorf("plan is not executable, %v", r)
                    fmt.Println(string(debug.Stack()))
                }
            }()

            // If plan is not executable, 1 of the computer will panic
            e.computers[computerOutputName].Compute(plan)
        }()
    }

    return
}

func (e Engine) Execute(ctx context.Context, planName string, plan any) error {
    computers, ok := e.plans[planName]
    if !ok {
        panic(ErrAnalyzePlanNotDone)
    }

    g, groupCtx := errgroup.WithContext(ctx)
    for _, computerOutputName := range computers {
        name := computerOutputName
        g.Go(func() error {
            task := e.computers[name].Compute(plan)
            task.Execute(groupCtx)

            return task.Error()
        })
    }

    return g.Wait()
}