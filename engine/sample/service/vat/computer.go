package vat

import (
	"context"

	"github.com/grab/async/async"
	"github.com/grab/async/engine/sample/config"
)

// Computers without any external dependencies can register itself directly
// with the engine using init()
func init() {
	// config.Print("vat")
	config.Engine.RegisterComputer(Amount{}, computer{})
	// config.Print(config.Engine)
}

type computer struct{}

func (c computer) Compute(p any) async.SilentTask {
	casted := p.(plan)

	task := async.NewSilentTask(
		func(ctx context.Context) error {
			c.addVATAmount(casted)
			return nil
		},
	)

	return task
}
