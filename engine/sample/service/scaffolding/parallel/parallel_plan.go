package parallel

import (
	"github.com/grab/async/engine/sample/service/costconfigs"
	"github.com/grab/async/engine/sample/service/miscellaneous"
	"github.com/grab/async/engine/sample/service/scaffolding/sequential"
	"github.com/grab/async/engine/sample/service/travelcost"
	"github.com/grab/async/engine/sample/service/travelplan"
)

type ParallelPlan struct {
	miscellaneous.CostRequest
	costconfigs.CostConfigs
	travelplan.TravelPlan
	travelcost.TravelCost
	sequential.SequentialPlan
}

func NewPlan(r miscellaneous.CostRequest) *ParallelPlan {
	return &ParallelPlan{
		CostRequest: r,
	}
}

func (p *ParallelPlan) IsSequential() bool {
	return false
}
