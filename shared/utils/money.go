package utils

import "math"

func Round(value float64, decimal int) float64 {
	factor := math.Pow10(decimal)
	return math.Round(value*factor) / factor
}

func RoundUp(value float64, decimal int) float64 {
	factor := math.Pow10(decimal)
	return math.Ceil(value*factor) / factor
}

func RoundDown(value float64, decimal int) float64 {
	factor := math.Pow10(decimal)
	return math.Floor(value*factor) / factor
}

func RoundMoneyVND(v float64) float64 {
	return Round(v, 0)
}

func RoundVAT(v float64) float64 {
	return Round(v, 2)
}

func RoundCost(v float64) float64 {
	return Round(v, 4)
}
