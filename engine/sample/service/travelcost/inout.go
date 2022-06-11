package travelcost

import (
	"github.com/grab/async/engine/core"
)

type plan interface {
	input
	output
}

type input interface {
	GetBaseCost() float64
	GetTravelDistance() float64
	GetTravelDuration() float64
	GetCostPerKilometer() float64
	GetCostPerMinute() float64
}

type output interface {
	SetTravelCost(core.AsyncResult)
}

type TravelCost core.AsyncResult

func (r TravelCost) GetTravelCost() float64 {
	result := core.Extract[float64](r.Task)
	return result
}
