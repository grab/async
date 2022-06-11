package costconfigs

import (
	"github.com/grab/async/engine/core"
	"github.com/grab/async/engine/sample/service/costconfigs/dummy"
)

type plan interface {
	output
}

type output interface {
	SetCostConfigs(CostConfigs)
}

type CostConfigs core.AsyncResult

func (r CostConfigs) GetBaseCost() float64 {
	result := core.Take[dummy.MergedCostConfigs](r.Task)
	return result.BaseCost
}

func (r CostConfigs) GetCostPerKilometer() float64 {
	result := core.Take[dummy.MergedCostConfigs](r.Task)
	return result.CostPerKilometer
}

func (r CostConfigs) GetCostPerMinute() float64 {
	result := core.Take[dummy.MergedCostConfigs](r.Task)
	return result.CostPerMinute
}

func (r CostConfigs) GetPlatformFee() float64 {
	result := core.Take[dummy.MergedCostConfigs](r.Task)
	return result.PlatformFee
}

func (r CostConfigs) GetVATPercent() float64 {
	result := core.Take[dummy.MergedCostConfigs](r.Task)
	return result.VATPercent
}
