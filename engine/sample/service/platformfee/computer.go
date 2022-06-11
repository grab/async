package platformfee

import (
	"context"

	"github.com/grab/async/engine/sample/config"
)

// Computers without any external dependencies can register itself directly
// with the engine using init()
func init() {
	// config.Print("platformfee")
	config.Engine.RegisterSilentComputer(PlatformFee{}, computer{})
	// config.Print(config.Engine)
}

type computer struct{}

func (c computer) Compute(ctx context.Context, p any) error {
	casted := p.(plan)

	c.addPlatformFee(casted)

	return nil
}
