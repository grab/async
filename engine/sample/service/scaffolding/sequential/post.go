package sequential

import "fmt"

type post interface {
    GetTotalCost() float64
}

type postHook struct {}

func (postHook) PostExecute(p any) error {
    fmt.Println("After executing sequential plan")
    casted := p.(post)

    fmt.Println("Calculated total cost:", casted.GetTotalCost())

    return nil
}

type anotherPostHook struct {}

func (anotherPostHook) PostExecute(p any) error {
    fmt.Println("After sequential plan 2nd hook")

    return nil
}
