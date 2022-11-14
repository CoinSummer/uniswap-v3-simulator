package uniswap_v3_simulator

import "github.com/shopspring/decimal"

var (
	MaxUint128 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(128)).Sub(decimal.NewFromInt(1))
	MaxUint160 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(160)).Sub(decimal.NewFromInt(1))
	MaxUint256 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(256)).Sub(decimal.NewFromInt(1))
	MaxInt128  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(127)).Sub(decimal.NewFromInt(1))
	MinInt128  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(127)).Neg()

	Q32  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(32))
	Q96  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(96))
	Q128 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(128))
	Q192 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(192))

	MAX_FEE = decimal.NewFromInt(10).Pow(decimal.NewFromInt(6))

	TICK_SPACINGS = map[FeeAmount]int{
		FeeAmountLow:    10,
		FeeAmountMedium: 60,
		FeeAmountHigh:   200,
	}
	MIN_TICK          int = -887272
	MAX_TICK          int = -MIN_TICK
	MIN_SQRT_RATIO        = decimal.NewFromInt(4295128739)
	MAX_SQRT_RATIO, _     = decimal.NewFromString("1461446703485210103287273052203988822378723970342")

	ZERO = decimal.Zero
	ONE  = decimal.NewFromInt(1)
)
