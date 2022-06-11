package sequential

import (
	"github.com/grab/async/engine/sample/service/platformfee"
	"github.com/grab/async/engine/sample/service/vat"
)

type SequentialPlan struct {
	preHook
	totalCost float64
	platformfee.PlatformFee
	vat.Amount
	postHook
	anotherPostHook
}

func (p *SequentialPlan) IsSequential() bool {
	return true
}
