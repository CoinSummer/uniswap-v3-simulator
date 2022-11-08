package uniswap_v3_simulator

import (
	"errors"
	"github.com/shopspring/decimal"
)

var OVERFLOW = errors.New("OVERFLOW")
var UNDERFLOW = errors.New("UNDERFLOW")

func LiquidityAddDelta(x decimal.Decimal, y decimal.Decimal) (decimal.Decimal, error) {
	if x.GreaterThan(MaxUint128) {
		return decimal.Zero, OVERFLOW
	}
	if y.GreaterThan(MaxUint128) {
		return decimal.Zero, OVERFLOW
	}
	if y.IsNegative() {
		negy := y.Neg()
		if negy.GreaterThan(x) {
			return decimal.Zero, UNDERFLOW
		}
		return x.Sub(negy), nil
	} else {
		if x.Add(y).GreaterThan(MaxUint128) {
			return decimal.Zero, OVERFLOW
		}
		return x.Add(y), nil
	}
}
func AddDelta(x, y decimal.Decimal) (decimal.Decimal, error) {
	if x.GreaterThan(MaxUint128) || y.GreaterThanOrEqual(MaxUint128) {
		return decimal.Zero, OVERFLOW
	}
	if y.IsNegative() {
		negatedY := y.Neg()
		if !x.GreaterThanOrEqual(negatedY) {
			return decimal.Zero, UNDERFLOW
		}
		return x.Sub(negatedY), nil
	} else {
		if !x.Add(y).LessThanOrEqual(MaxUint128) {
			return decimal.Zero, OVERFLOW
		}
		return x.Add(y), nil

	}
}
