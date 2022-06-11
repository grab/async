package platformfee

import (
	"context"

	"github.com/grab/async/async"
	"github.com/grab/async/engine/sample/config"
)

// Computers without any external dependencies can register itself directly
// with the engine using init()
func init() {
	// fmt.Println("platformfee")
	config.Engine.RegisterComputer(PlatformFee{}, computer{})
	// fmt.Println(config.Engine)
}

type computer struct{}

func (c computer) Compute(p any) async.SilentTask {
	casted := p.(plan)

	task := async.NewSilentTask(
		func(ctx context.Context) error {
			c.addPlatformFee(casted)
			return nil
		},
	)

	return task
}
