package parallel

import (
	"context"

	"github.com/grab/async/engine/sample/config"
	"github.com/grab/async/engine/sample/service/costconfigs"
	"github.com/grab/async/engine/sample/service/travelcost"
	"github.com/grab/async/engine/sample/service/travelplan"
)

var planName string

func init() {
	// config.Print("ParallelPlan")
	planName = config.Engine.AnalyzePlan(&ParallelPlan{})
}

func (p *ParallelPlan) SetCostConfigs(o costconfigs.CostConfigs) {
	p.CostConfigs = o
}

func (p *ParallelPlan) SetTravelPlan(o travelplan.TravelPlan) {
	p.TravelPlan = o
}

func (p *ParallelPlan) SetTravelCost(o travelcost.TravelCost) {
	p.TravelCost = o
}

func (p *ParallelPlan) Execute(ctx context.Context) error {
	return config.Engine.Execute(ctx, planName, p)
}
