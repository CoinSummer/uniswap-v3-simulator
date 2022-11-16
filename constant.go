package uniswap_v3_simulator

import (
	"github.com/shopspring/decimal"
	"math/big"
)

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

var (
	sqrtConst1, _  = new(big.Int).SetString("fffcb933bd6fad37aa2d162d1a594001", 16)
	sqrtConst2, _  = new(big.Int).SetString("100000000000000000000000000000000", 16)
	sqrtConst3, _  = new(big.Int).SetString("fff97272373d413259a46990580e213a", 16)
	sqrtConst4, _  = new(big.Int).SetString("fff2e50f5f656932ef12357cf3c7fdcc", 16)
	sqrtConst5, _  = new(big.Int).SetString("ffe5caca7e10e4e61c3624eaa0941cd0", 16)
	sqrtConst6, _  = new(big.Int).SetString("ffcb9843d60f6159c9db58835c926644", 16)
	sqrtConst7, _  = new(big.Int).SetString("ff973b41fa98c081472e6896dfb254c0", 16)
	sqrtConst8, _  = new(big.Int).SetString("ff2ea16466c96a3843ec78b326b52861", 16)
	sqrtConst9, _  = new(big.Int).SetString("fe5dee046a99a2a811c461f1969c3053", 16)
	sqrtConst10, _ = new(big.Int).SetString("fcbe86c7900a88aedcffc83b479aa3a4", 16)
	sqrtConst11, _ = new(big.Int).SetString("f987a7253ac413176f2b074cf7815e54", 16)
	sqrtConst12, _ = new(big.Int).SetString("f3392b0822b70005940c7a398e4b70f3", 16)
	sqrtConst13, _ = new(big.Int).SetString("e7159475a2c29b7443b29c7fa6e889d9", 16)
	sqrtConst14, _ = new(big.Int).SetString("d097f3bdfd2022b8845ad8f792aa5825", 16)
	sqrtConst15, _ = new(big.Int).SetString("a9f746462d870fdf8a65dc1f90e061e5", 16)
	sqrtConst16, _ = new(big.Int).SetString("70d869a156d2a1b890bb3df62baf32f7", 16)
	sqrtConst17, _ = new(big.Int).SetString("31be135f97d08fd981231505542fcfa6", 16)
	sqrtConst18, _ = new(big.Int).SetString("9aa508b5b7a84e1c677de54f3e99bc9", 16)
	sqrtConst19, _ = new(big.Int).SetString("5d6af8dedb81196699c329225ee604", 16)
	sqrtConst20, _ = new(big.Int).SetString("2216e584f5fa1ea926041bedfe98", 16)
	sqrtConst21, _ = new(big.Int).SetString("48a170391f7dc42444e8fa2", 16)
)
