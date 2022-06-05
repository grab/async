package travelcost

func (computer) doCalculation(in input) float64 {
    return in.GetBaseCost() + in.GetTravelDuration() * in.GetCostPerMinute() + in.GetTravelDistance() * in.GetCostPerKilometer()
}