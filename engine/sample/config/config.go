package config

import "github.com/grab/async/engine/core"

var Engine = &CostEngine{
	Engine: core.NewEngine(),
}

type CostEngine struct {
	core.Engine
}
