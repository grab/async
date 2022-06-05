package main

import (
    "context"
    "fmt"

    "github.com/grab/async/engine/sample/server"
    "github.com/grab/async/engine/sample/service"
)

func main() {
    server.Serve()

    var p service.ConcretePlan
    if err := p.Execute(context.Background()) ; err != nil {
        fmt.Println(err)
    }

    fmt.Println(p.GetTravelCost())
}
