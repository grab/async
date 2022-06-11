package dummy

import (
	"github.com/grab/async/engine/sample/config"
)

var planName string

func init() {
	// fmt.Println("DummyPlan")
	planName = config.Engine.AnalyzePlan(&DummyPlan{})
}

func (p *DummyPlan) SetSomething(something float64) {
	p.something = something
}