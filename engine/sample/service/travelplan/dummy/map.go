package dummy

import (
	"github.com/grab/async/engine/sample/config"
	"github.com/stretchr/testify/assert"
)

type TravelPlan struct {
	Distance float64
	Duration float64
}

type MapService struct{}

func (MapService) BuildTravelPlan(pointA, pointB string) (TravelPlan, error) {
	config.Printf("Building travel plan from %s to %s using real map\n", pointA, pointB)
	return TravelPlan{
		Distance: 2,
		Duration: 3,
	}, assert.AnError
}
