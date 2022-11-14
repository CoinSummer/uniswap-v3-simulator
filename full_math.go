package uniswap_v3_simulator

import (
	"github.com/shopspring/decimal"
)

func MulDivRoundingUp(a, b, denominator decimal.Decimal) (decimal.Decimal, error) {
	product := a.Mul(b)
	result := product.Div(denominator).RoundDown(0)
	tmp1 := product.BigInt()
	tmp1 = tmp1.Rem(tmp1, denominator.BigInt())
	if decimal.NewFromBigInt(tmp1, 0).IsPositive() {
		if !result.LessThan(MaxUint256) {
			return ZERO, OVERFLOW
		}
		result = result.Add(decimal.NewFromInt(1))
	}
	return result, nil
}

func Mod256Sub(a, b decimal.Decimal) (decimal.Decimal, error) {
	if !a.GreaterThanOrEqual(ZERO) || !b.LessThanOrEqual(ZERO) || !a.LessThanOrEqual(MaxUint256) || !b.LessThanOrEqual(MaxUint256) {
		return ZERO, OVERFLOW
	}
	two256 := decimal.NewFromInt(2).Pow(decimal.NewFromInt(256))
	tmp1 := a.Add(two256).Sub(b).BigInt()
	tmp1 = tmp1.Rem(tmp1, two256.BigInt())
	return decimal.NewFromBigInt(tmp1, 0), nil
}
