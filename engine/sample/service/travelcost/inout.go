package travelcost

import "github.com/grab/async/async"

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
    SetTravelCost(TravelCost)
}

type TravelCost struct {
    task async.Task[float64]
}

func (r TravelCost) GetTravelCost() float64 {
    result, _ := r.task.Outcome()
    return result
}
