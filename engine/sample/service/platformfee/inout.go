package platformfee

type plan interface {
    input
    output
}

type input interface {
    GetPlatformFee() float64
    GetTotalCost() float64
}

type output interface {
    SetTotalCost(float64)
}

type PlatformFee struct {}
