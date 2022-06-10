package dummy

import (
    "context"

    "github.com/grab/async/engine/sample/config"
)

var planName string

func init() {
    // fmt.Println("DummyPlan")
    planName = config.Engine.AnalyzePlan(&DummyPlan{})
}

func (p *DummyPlan) SetSomething(something float64) {
    p.something = something
}

func (p *DummyPlan) Execute(ctx context.Context) error {
    return config.Engine.Execute(ctx, planName, p)
}