package main

import (
	"context"
	"fmt"

	"github.com/grab/async/engine/sample/config"
	"github.com/grab/async/engine/sample/server"
	"github.com/grab/async/engine/sample/service/miscellaneous"
	"github.com/grab/async/engine/sample/service/scaffolding/parallel"
	"github.com/grab/async/engine/sample/service/scaffolding/sequential"
)

type customPostHook struct{}

func (customPostHook) PostExecute(p any) error {
	fmt.Println("After sequential plan custom hook")

	return nil
}

func main() {
	server.Serve()

	config.Engine.ConnectPostHook(&sequential.SequentialPlan{}, customPostHook{})

	p := parallel.NewPlan(
		miscellaneous.CostRequest{
			PointA: "Clementi",
			PointB: "Changi Airport",
		},
	)

	if err := p.Execute(context.Background()); err != nil {
		fmt.Println(err)
	}

	fmt.Println(p.GetTravelCost())
	fmt.Println(p.GetTotalCost())
}
