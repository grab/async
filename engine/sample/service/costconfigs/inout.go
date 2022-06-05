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

func (r CostConfigs) GetTravelDistance() float64 {
    result, _ := r.task.Outcome()
    return result.TravelDistance
}

func (r CostConfigs) GetTravelDuration() float64 {
    result, _ := r.task.Outcome()
    return result.TravelDuration
}

func (r CostConfigs) GetCostPerKilometer() float64 {
    result, _ := r.task.Outcome()
    return result.CostPerKilometer
}

func (r CostConfigs) GetCostPerMinute() float64 {
    result, _ := r.task.Outcome()
    return result.CostPerMinute
}
