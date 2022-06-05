package travelplan

import (
	"github.com/grab/async/async"
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
	SetTravelPlan(TravelPlan)
}

type TravelPlan struct {
	task async.Task[dummy.TravelPlan]
}

func (p TravelPlan) GetTravelDistance() float64 {
	result, _ := p.task.Outcome()
	return result.Distance
}

func (p TravelPlan) GetTravelDuration() float64 {
	result, _ := p.task.Outcome()
	return result.Duration
}
