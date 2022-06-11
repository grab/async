package costconfigs

import (
	"github.com/grab/async/async"
	"github.com/grab/async/engine/sample/service/costconfigs/dummy"
)

type plan interface {
	output
}

type output interface {
	SetCostConfigs(CostConfigs)
}

type CostConfigs struct {
	task async.Task[dummy.MergedCostConfigs]
}

func (r CostConfigs) GetBaseCost() float64 {
	result, _ := r.task.Outcome()
	return result.BaseCost
}

func (r CostConfigs) GetCostPerKilometer() float64 {
	result, _ := r.task.Outcome()
	return result.CostPerKilometer
}

func (r CostConfigs) GetCostPerMinute() float64 {
	result, _ := r.task.Outcome()
	return result.CostPerMinute
}

func (r CostConfigs) GetPlatformFee() float64 {
	result, _ := r.task.Outcome()
	return result.PlatformFee
}

func (r CostConfigs) GetVATPercent() float64 {
	result, _ := r.task.Outcome()
	return result.VATPercent
}
