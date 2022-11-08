package uniswap_v3_simulator

import (
	"github.com/shopspring/decimal"
)

func MulDivRoundingUp(a, b, denominator decimal.Decimal) (decimal.Decimal, error) {

}
func Mod256Sub(a, b decimal.Decimal) (decimal.Decimal, error) {
	if !a.GreaterThanOrEqual(decimal.Zero) || !b.LessThanOrEqual(decimal.Zero) || !a.LessThanOrEqual(MaxUint256) || !b.LessThanOrEqual(MaxUint256) {
		return decimal.Zero, OVERFLOW
	}
	two256 := decimal.NewFromInt(2).Pow(decimal.NewFromInt(256))
	tmp1 := a.Add(two256).Sub(b).BigInt()
	tmp1 = tmp1.Rem(tmp1, two256.BigInt())
	return decimal.NewFromBigInt(tmp1, 0), nil
}
