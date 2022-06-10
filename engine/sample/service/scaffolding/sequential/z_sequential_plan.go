package sequential

import (
    "context"

    "github.com/grab/async/engine/sample/config"
)

var planName string

func init() {
    // fmt.Println("SequentialPlan")
    planName = config.Engine.AnalyzePlan(&SequentialPlan{})
}

func (p *SequentialPlan) GetTotalCost() float64 {
    return p.totalCost
}

func (p *SequentialPlan) SetTotalCost(totalCost float64) {
    p.totalCost = totalCost
}

func (p *SequentialPlan) Execute(ctx context.Context) error {
    return config.Engine.Execute(ctx, planName, p)
}