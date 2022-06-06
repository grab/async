package platformfee

import (
    "context"

    "github.com/grab/async/engine/sample/config"
)

// Computers without any external dependencies can register itself directly
// with the engine using init()
func init() {
    // fmt.Println("platformfee")
    config.Engine.RegisterSyncComputer(PlatformFee{}, computer{})
    // fmt.Println(config.Engine)
}

type computer struct{}

func (c computer) Compute(ctx context.Context, p any) error {
    casted := p.(plan)

    c.addPlatformFee(casted)

    return nil
}

