package dummy

type MergedCostConfigs struct {
    BaseCost       float64
    TravelDistance float64
    TravelDuration   float64
    CostPerKilometer float64
    CostPerMinute    float64
}

type CostConfigsFetcher struct {}

func (CostConfigsFetcher) Fetch() MergedCostConfigs {
    return MergedCostConfigs{
        BaseCost:         1,
        TravelDistance:   2,
        TravelDuration:   3,
        CostPerKilometer: 4,
        CostPerMinute:    5,
    }
}
