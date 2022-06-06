package vat

func (computer) addVATAmount(p plan) {
    vatAmount := p.GetTotalCost() * p.GetVATPercent() / 100
    p.SetTotalCost(p.GetTotalCost() + vatAmount)
}

