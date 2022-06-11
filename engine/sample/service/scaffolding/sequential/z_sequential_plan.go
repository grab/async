package sequential

import (
	"github.com/grab/async/engine/sample/config"
)

var planName string

func init() {
	// config.Print("SequentialPlan")
	planName = config.Engine.AnalyzePlan(&SequentialPlan{})
}

func (p *SequentialPlan) GetTotalCost() float64 {
	return p.totalCost
}

func (p *SequentialPlan) SetTotalCost(totalCost float64) {
	p.totalCost = totalCost
}
