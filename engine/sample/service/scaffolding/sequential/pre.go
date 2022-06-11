package sequential

import (
	"fmt"
)

type pre interface {
	GetTravelCost() float64
	SetTotalCost(float64)
}

type preHook struct{}

func (preHook) PreExecute(p any) error {
	fmt.Println("Before executing sequential plan")
	casted := p.(pre)

	casted.SetTotalCost(casted.GetTravelCost())

	return nil
}