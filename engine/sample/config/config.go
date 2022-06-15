package config

import (
	"fmt"

	"github.com/grab/async/engine/core"
)

var Engine = &CostEngine{
	Engine: core.NewEngine(),
}

var printDebugLog = true

func Print(values ...any) {
	if printDebugLog {
		fmt.Println(values...)
	}
}

func Printf(format string, values ...any) {
	if printDebugLog {
		fmt.Printf(format, values...)
	}
}

type CostEngine struct {
	core.Engine
	// Add common utilities like logger, statsD, UCM client, etc.
	// for all component codes to share.
}
