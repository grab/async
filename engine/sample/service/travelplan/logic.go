package travelplan

import "github.com/grab/async/engine/sample/service/travelplan/dummy"

func (c computer) buildTravelPlan(p plan) (dummy.TravelPlan, error) {
	return c.mapService.BuildTravelPlan(p.GetPointA(), p.GetPointB())
}
