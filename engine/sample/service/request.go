package service

type CostRequest struct {
	PointA string
	PointB string
}

func (r CostRequest) GetPointA() string {
	return r.PointA
}

func (r CostRequest) GetPointB() string {
	return r.PointB
}
