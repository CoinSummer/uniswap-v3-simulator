package uniswap_v3_simulator

import (
	"github.com/shopspring/decimal"
)

var (
	MaxUint128 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(128)).Sub(decimal.NewFromInt(1))
	MaxUint256 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(256)).Sub(decimal.NewFromInt(1))
	MaxInt128  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(127)).Sub(decimal.NewFromInt(1))
	MinInt128  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(127)).Neg()

	Q32  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(32))
	Q96  = decimal.NewFromInt(2).Pow(decimal.NewFromInt(96))
	Q128 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(128))

	MIN_TICK          int = -887272
	MAX_TICK          int = -MIN_TICK
	MIN_SQRT_RATIO        = decimal.NewFromInt(4295128739)
	MAX_SQRT_RATIO, _     = decimal.NewFromString("1461446703485210103287273052203988822378723970342")

	ZERO = decimal.NewFromInt(0)
	ONE  = decimal.NewFromInt(1)
)
