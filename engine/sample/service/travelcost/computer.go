package travelcost

import (
	"context"

	"github.com/grab/async/async"
	"github.com/grab/async/engine/sample/config"
)

// Computers without any external dependencies can register itself directly
// with the engine using init()
func init() {
	// fmt.Println("travelcost")
	config.Engine.RegisterComputer(TravelCost{}, computer{})
	// fmt.Println(config.Engine)
}

type computer struct{}

func (c computer) Compute(p any) async.SilentTask {
	casted := p.(plan)

	task := async.NewTask(
		func(ctx context.Context) (float64, error) {
			return c.doCalculation(casted), nil
		},
	)

	casted.SetTravelCost(
		TravelCost{
			task: task,
		},
	)

	return task
}
