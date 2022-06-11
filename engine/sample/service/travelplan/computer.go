package travelplan

import (
	"context"

	"github.com/grab/async/async"
	"github.com/grab/async/engine/sample/config"
	"github.com/grab/async/engine/sample/service/travelplan/dummy"
)

// Computers with external dependencies still has to register itself with the
// engine using init() so that we can perform validations on plans
func init() {
	// config.Print("travelplan")
	config.Engine.RegisterComputer(TravelPlan{}, computer{})
	// config.Print(config.Engine)
}

type computer struct {
	mapService dummy.MapService
}

// Computers with external dependencies can register itself with the engine
// via an exported InitComputer() that takes in dependencies as arguments
// to overwrite the dummy computer registered via init()

// InitComputer ...
func InitComputer(mapService dummy.MapService) {
	c := computer{
		mapService: mapService,
	}

	// config.Print("travelplan")
	config.Engine.RegisterComputer(TravelPlan{}, c)
	// config.Print(config.Engine)
}

func (c computer) Compute(p any) async.SilentTask {
	casted := p.(plan)

	task := async.NewTask(
		func(ctx context.Context) (dummy.TravelPlan, error) {
			travelPlan, err := c.buildTravelPlan(casted)
			if err != nil {
				return c.calculateStraightLineDistance(casted), nil
			}

			return travelPlan, nil
		},
	)

	casted.SetTravelPlan(
		TravelPlan{
			task: task,
		},
	)

	return task
}
