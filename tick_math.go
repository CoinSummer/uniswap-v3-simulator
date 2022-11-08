package uniswap_v3_simulator

import "github.com/shopspring/decimal"

func TickSpacingToMaxLiquidityPerTick(tickSpacing int64) decimal.Decimal {
	ts := decimal.NewFromInt(tickSpacing)
	minTick := decimal.NewFromInt(MIN_TICK).Div(ts).Floor().Mul(ts)
	maxTick := decimal.NewFromInt(MAX_TICK).Div(ts).Floor().Mul(ts)
	numTicks := maxTick.Sub(minTick).Div(ts).Floor().Add(decimal.NewFromInt(1))
	return MaxUint128.Div(numTicks).Floor()
}
