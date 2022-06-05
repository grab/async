package service

import (
    "context"

    "github.com/grab/async/engine/sample/config"
    "github.com/grab/async/engine/sample/service/costconfigs"
    "github.com/grab/async/engine/sample/service/travelcost"
)

var planName string

func init() {
    // fmt.Println("ConcretePlan")
    planName = config.Engine.AnalyzePlan(ConcretePlan{})
}

type ConcretePlan struct {
    travelcost.TravelCost
    costconfigs.CostConfigs
}

func (c *ConcretePlan) SetTravelCost(o travelcost.TravelCost) {
    c.TravelCost = o
}

func (c *ConcretePlan) SetCostConfigs(o costconfigs.CostConfigs) {
    c.CostConfigs = o
}

func (c *ConcretePlan) Execute(ctx context.Context) error {
    return config.Engine.Execute(ctx, planName, c)
}
