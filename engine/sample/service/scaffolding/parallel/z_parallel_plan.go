package parallel

import (
	"github.com/grab/async/engine/core"
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

func (p *ParallelPlan) SetCostConfigs(o core.AsyncResult) {
	p.CostConfigs = (costconfigs.CostConfigs)(o)
}

func (p *ParallelPlan) SetTravelPlan(o travelplan.TravelPlan) {
	p.TravelPlan = o
}

func (p *ParallelPlan) SetTravelCost(o travelcost.TravelCost) {
	p.TravelCost = o
}
