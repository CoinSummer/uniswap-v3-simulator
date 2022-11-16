package uniswap_v3_simulator

import (
	"errors"
	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
)

var POWERS_OF_2 []struct {
	i   int
	pow decimal.Decimal
}

func init() {
	for _, i := range []int{128, 64, 32, 16, 8, 4, 2, 1} {
		POWERS_OF_2 = append(POWERS_OF_2, struct {
			i   int
			pow decimal.Decimal
		}{
			i:   i,
			pow: decimal.NewFromInt(2).Pow(decimal.NewFromInt(int64(i))),
		})
	}

}
func TickSpacingToMaxLiquidityPerTick(tickSpacing int) decimal.Decimal {
	ts := decimal.NewFromInt(int64(tickSpacing))
	minTick := decimal.NewFromInt(int64(MIN_TICK)).Div(ts).RoundDown(0).Mul(ts)
	maxTick := decimal.NewFromInt(int64(MAX_TICK)).Div(ts).RoundDown(0).Mul(ts)
	numTicks := maxTick.Sub(minTick).Div(ts).RoundDown(0).Add(decimal.NewFromInt(1))
	return MaxUint128.Div(numTicks).RoundDown(0)
}

var (
	ErrInvalidTick      = errors.New("invalid tick")
	ErrInvalidSqrtRatio = errors.New("invalid sqrt ratio")
	magicSqrt10001, _   = new(big.Int).SetString("255738958999603826347141", 10)
	magicTickLow, _     = new(big.Int).SetString("3402992956809132418596140100660247210", 10)
	magicTickHigh, _    = new(big.Int).SetString("291339464771989622907027621153398088495", 10)
)

func GetTickAtSqrtRatio(sqrtRatioX96D decimal.Decimal) (int, error) {
	sqrtRatioX96 := sqrtRatioX96D.BigInt()

	if sqrtRatioX96.Cmp(MIN_SQRT_RATIO.BigInt()) < 0 || sqrtRatioX96.Cmp(MAX_SQRT_RATIO.BigInt()) >= 0 {
		return 0, ErrInvalidSqrtRatio
	}
	sqrtRatioX128 := new(big.Int).Lsh(sqrtRatioX96, 32)
	msb, err := MostSignificantBit(sqrtRatioX128)
	if err != nil {
		return 0, err
	}
	var r *big.Int
	if big.NewInt(msb).Cmp(big.NewInt(128)) >= 0 {
		r = new(big.Int).Rsh(sqrtRatioX128, uint(msb-127))
	} else {
		r = new(big.Int).Lsh(sqrtRatioX128, uint(127-msb))
	}

	log2 := new(big.Int).Lsh(new(big.Int).Sub(big.NewInt(msb), big.NewInt(128)), 64)

	for i := 0; i < 14; i++ {
		r = new(big.Int).Rsh(new(big.Int).Mul(r, r), 127)
		f := new(big.Int).Rsh(r, 128)
		log2 = new(big.Int).Or(log2, new(big.Int).Lsh(f, uint(63-i)))
		r = new(big.Int).Rsh(r, uint(f.Int64()))
	}

	logSqrt10001 := new(big.Int).Mul(log2, magicSqrt10001)

	tickLow := new(big.Int).Rsh(new(big.Int).Sub(logSqrt10001, magicTickLow), 128).Int64()
	tickHigh := new(big.Int).Rsh(new(big.Int).Add(logSqrt10001, magicTickHigh), 128).Int64()

	if tickLow == tickHigh {
		return int(tickLow), nil
	}

	sqrtRatio, err := GetSqrtRatioAtTick(int(tickHigh))
	if err != nil {
		return 0, err
	}
	if sqrtRatio.BigInt().Cmp(sqrtRatioX96) <= 0 {
		return int(tickHigh), nil
	} else {
		return int(tickLow), nil
	}
}

func mulShift(val decimal.Decimal, mulBy string) decimal.Decimal {
	byBi, _ := big.NewInt(0).SetString(mulBy, 16)
	by := decimal.NewFromBigInt(byBi, 0)
	tmp := val.Mul(by).BigInt()
	tmp = tmp.Rsh(tmp, 128)
	return decimal.NewFromBigInt(tmp, 0)
}

var (
	INVALID_TICK = errors.New("invalid tick")
)

func GetSqrtRatioAtTick(tick int) (decimal.Decimal, error) {
	if tick < MIN_TICK || tick > MAX_TICK {
		return ZERO, INVALID_TICK
	}
	var absTick = int(math.Abs(float64(tick)))
	var ratio decimal.Decimal
	var ratioBi *big.Int
	if absTick&0x1 != 0 {
		ratioBi, _ = big.NewInt(0).SetString("fffcb933bd6fad37aa2d162d1a594001", 16)
	} else {
		ratioBi, _ = big.NewInt(0).SetString("100000000000000000000000000000000", 16)
	}
	ratio = decimal.NewFromBigInt(ratioBi, 0)

	if (absTick & 0x2) != 0 {
		ratio = mulShift(ratio, "fff97272373d413259a46990580e213a")
	}
	if (absTick & 0x4) != 0 {

		ratio = mulShift(ratio, "fff2e50f5f656932ef12357cf3c7fdcc")
	}
	if (absTick & 0x8) != 0 {

		ratio = mulShift(ratio, "ffe5caca7e10e4e61c3624eaa0941cd0")
	}
	if (absTick & 0x10) != 0 {

		ratio = mulShift(ratio, "ffcb9843d60f6159c9db58835c926644")
	}
	if (absTick & 0x20) != 0 {
		ratio = mulShift(ratio, "ff973b41fa98c081472e6896dfb254c0")

	}
	if (absTick & 0x40) != 0 {

		ratio = mulShift(ratio, "ff2ea16466c96a3843ec78b326b52861")
	}
	if (absTick & 0x80) != 0 {

		ratio = mulShift(ratio, "fe5dee046a99a2a811c461f1969c3053")
	}
	if (absTick & 0x100) != 0 {

		ratio = mulShift(ratio, "fcbe86c7900a88aedcffc83b479aa3a4")
	}
	if (absTick & 0x200) != 0 {

		ratio = mulShift(ratio, "f987a7253ac413176f2b074cf7815e54")
	}
	if (absTick & 0x400) != 0 {

		ratio = mulShift(ratio, "f3392b0822b70005940c7a398e4b70f3")
	}
	if (absTick & 0x800) != 0 {

		ratio = mulShift(ratio, "e7159475a2c29b7443b29c7fa6e889d9")
	}
	if (absTick & 0x1000) != 0 {

		ratio = mulShift(ratio, "d097f3bdfd2022b8845ad8f792aa5825")
	}
	if (absTick & 0x2000) != 0 {

		ratio = mulShift(ratio, "a9f746462d870fdf8a65dc1f90e061e5")
	}
	if (absTick & 0x4000) != 0 {

		ratio = mulShift(ratio, "70d869a156d2a1b890bb3df62baf32f7")
	}
	if (absTick & 0x8000) != 0 {

		ratio = mulShift(ratio, "31be135f97d08fd981231505542fcfa6")
	}
	if (absTick & 0x10000) != 0 {

		ratio = mulShift(ratio, "9aa508b5b7a84e1c677de54f3e99bc9")
	}
	if (absTick & 0x20000) != 0 {
		ratio = mulShift(ratio, "5d6af8dedb81196699c329225ee604")
	}
	if (absTick & 0x40000) != 0 {

		ratio = mulShift(ratio, "2216e584f5fa1ea926041bedfe98")
	}
	if (absTick & 0x80000) != 0 {

		ratio = mulShift(ratio, "48a170391f7dc42444e8fa2")
	}
	if tick > 0 {
		ratio = MaxUint256.Div(ratio).RoundDown(0)
	}
	_, remainder := ratio.QuoRem(Q32, 0)
	remainder = remainder.RoundDown(0)
	if remainder.GreaterThan(ZERO) {
		return ratio.Div(Q32).Add(decimal.NewFromInt(1)).RoundDown(0), nil
	} else {
		return ratio.Div(Q32).RoundDown(0), nil
	}
}

var ErrInvalidInput = errors.New("invalid input")

func MostSignificantBit(x *big.Int) (int64, error) {
	if x.Cmp(constants.Zero) <= 0 {
		return 0, ErrInvalidInput
	}
	if x.Cmp(MaxUint256.BigInt()) > 0 {
		return 0, ErrInvalidInput
	}
	var msb int64
	for _, power := range []int64{128, 64, 32, 16, 8, 4, 2, 1} {
		min := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(power)), nil)
		if x.Cmp(min) >= 0 {
			x = new(big.Int).Rsh(x, uint(power))
			msb += power
		}
	}
	return msb, nil
}
