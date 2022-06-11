package travelplan

import (
	"github.com/grab/async/engine/core"
	"github.com/grab/async/engine/sample/service/travelplan/dummy"
)

type plan interface {
	input
	output
}

type input interface {
	GetPointA() string
	GetPointB() string
}

type output interface {
	SetTravelPlan(core.AsyncResult)
}

type TravelPlan core.AsyncResult

func (p TravelPlan) GetTravelDistance() float64 {
	result := core.Extract[dummy.TravelPlan](p.Task)
	return result.Distance
}

func (p TravelPlan) GetTravelDuration() float64 {
	result := core.Extract[dummy.TravelPlan](p.Task)
	return result.Duration
}
