package server

import (
    "github.com/grab/async/engine/sample/service/costconfigs"
    "github.com/grab/async/engine/sample/service/costconfigs/dummy"
)

func Serve() {
    costconfigs.InitComputer(dummy.CostConfigsFetcher{})
}