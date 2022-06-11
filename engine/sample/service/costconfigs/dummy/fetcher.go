package dummy

type MergedCostConfigs struct {
	BaseCost         float64
	CostPerKilometer float64
	CostPerMinute    float64
	PlatformFee      float64
	VATPercent       float64
}

type CostConfigsFetcher struct{}

func (CostConfigsFetcher) Fetch() MergedCostConfigs {
	return MergedCostConfigs{
		BaseCost:         1,
		CostPerKilometer: 4,
		CostPerMinute:    5,
		PlatformFee:      8,
		VATPercent:       10,
	}
}
