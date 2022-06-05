package service

import (
	"github.com/grab/async/engine/sample/service/costconfigs"
	"github.com/grab/async/engine/sample/service/travelcost"
	"github.com/grab/async/engine/sample/service/travelplan"
)

type ConcretePlan struct {
	CostRequest
	costconfigs.CostConfigs
	travelplan.TravelPlan
	travelcost.TravelCost
}

func NewPlan(r CostRequest) *ConcretePlan {
	return &ConcretePlan{
		CostRequest: r,
	}
}
func (c *ConcretePlan) IsSequential() bool {
	return false
}
