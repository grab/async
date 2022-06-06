package scaffolding

import (
	"context"
	"fmt"

	"github.com/grab/async/engine/sample/service/costconfigs"
	"github.com/grab/async/engine/sample/service/miscellaneous"
	"github.com/grab/async/engine/sample/service/travelcost"
	"github.com/grab/async/engine/sample/service/travelplan"
)

type ParallelPlan struct {
	miscellaneous.CostRequest
	costconfigs.CostConfigs
	travelplan.TravelPlan
	travelcost.TravelCost
}

func NewPlan(r miscellaneous.CostRequest) *ParallelPlan {
	return &ParallelPlan{
		CostRequest: r,
	}
}

func (c *ParallelPlan) IsSequential() bool {
	return false
}

func (c *ParallelPlan) PreExecute(ctx context.Context, masterPlan any) error {
	fmt.Println("Before executing parallel plan")
	return nil
}

func (c *ParallelPlan) PostExecute(ctx context.Context, masterPlan any) error {
	fmt.Println("After executing parallel plan")
	return nil
}
