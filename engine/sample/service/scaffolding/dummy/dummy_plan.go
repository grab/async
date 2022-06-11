package dummy

import (
	"github.com/grab/async/engine/sample/service/platformfee"
	"github.com/grab/async/engine/sample/service/vat"
)

type DummyPlan struct {
	something float64
	platformfee.PlatformFee
	vat.Amount
}

func (p *DummyPlan) IsSequential() bool {
	return true
}
