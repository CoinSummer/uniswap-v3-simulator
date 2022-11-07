package uniswap_v3_simulator

import "github.com/shopspring/decimal"

type FeeAmount int

const (
	FeeAmountLow    FeeAmount = 500
	FeeAmountMedium FeeAmount = 3000
	FeeAmountHigh   FeeAmount = 10000
)

type Pool struct {
	Token0               string
	Token1               string
	Fee                  FeeAmount
	TickSpacing          int64
	MaxLiquidityPerTick  decimal.Decimal
	Token0Balance        decimal.Decimal
	Token1Balance        decimal.Decimal
	SqrtPriceX96         decimal.Decimal
	Liquidity            decimal.Decimal
	TickCurrent          int64
	FeeGrowthGlobal0X128 decimal.Decimal
	FeeGrowthGlobal1X128 decimal.Decimal
	TickManager          TickManager
	PositionManager      PositionManager
}
