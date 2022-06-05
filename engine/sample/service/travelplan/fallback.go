package travelplan

import (
	"fmt"

	"github.com/grab/async/engine/sample/service/travelplan/dummy"
)

func (c computer) calculateStraightLineDistance(p plan) dummy.TravelPlan {
	fmt.Printf("Building travel plan from %s to %s using straight-line distance\n", p.GetPointA(), p.GetPointB())
	return dummy.TravelPlan{
		Distance: 4,
		Duration: 5,
	}
}
