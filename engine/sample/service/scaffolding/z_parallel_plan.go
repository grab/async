package scaffolding

import (
	"context"

	"github.com/grab/async/engine/sample/config"
	"github.com/grab/async/engine/sample/service/costconfigs"
	"github.com/grab/async/engine/sample/service/travelcost"
	"github.com/grab/async/engine/sample/service/travelplan"
)

var planName string

func init() {
	// fmt.Println("ParallelPlan")
	planName = config.Engine.AnalyzePlan(&ParallelPlan{})
}

func (c *ParallelPlan) SetCostConfigs(o costconfigs.CostConfigs) {
	c.CostConfigs = o
}

func (c *ParallelPlan) SetTravelPlan(o travelplan.TravelPlan) {
	c.TravelPlan = o
}

func (c *ParallelPlan) SetTravelCost(o travelcost.TravelCost) {
	c.TravelCost = o
}

func (c *ParallelPlan) Execute(ctx context.Context) error {
	return config.Engine.Execute(ctx, planName, c)
}
