package main

import (
	"context"

	"github.com/grab/async/engine/sample/config"
	"github.com/grab/async/engine/sample/server"
	"github.com/grab/async/engine/sample/service/miscellaneous"
	"github.com/grab/async/engine/sample/service/scaffolding/parallel"
	"github.com/grab/async/engine/sample/service/scaffolding/sequential"
)

type customPostHook struct{}

func (customPostHook) PostExecute(p any) error {
	config.Print("After sequential plan custom hook")

	return nil
}

type dummy struct {}

func (dummy) Do() {
	fmt.Println("hello")
}

func main() {
	// method, ok := reflect.ValueOf(dummy{}).Type().MethodByName("Do")
	// if ok {
	// 	fmt.Println(method)
	// }
	//
	// method.Func.Call([]reflect.Value{reflect.ValueOf(dummy{})})

	testEngine()
}

func testEngine() {
	server.Serve()

	config.Engine.ConnectPostHook(&sequential.SequentialPlan{}, customPostHook{})

	p := parallel.NewPlan(
		miscellaneous.CostRequest{
			PointA: "Clementi",
			PointB: "Changi Airport",
		},
	)

	if err := p.Execute(context.Background()); err != nil {
		config.Print(err)
	}

	config.Print(p.GetTravelCost())
	config.Print(p.GetTotalCost())
}
