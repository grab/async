package platformfee

func (computer) addPlatformFee(p plan) {
	p.SetTotalCost(p.GetTotalCost() + p.GetPlatformFee())
}
