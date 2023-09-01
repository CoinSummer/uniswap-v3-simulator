package uniswap_v3_simulator

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
)

func GetAmount1Delta(
	sqrtRatioAX96 decimal.Decimal,
	sqrtRatioBX96 decimal.Decimal,
	liquidity decimal.Decimal,
) (decimal.Decimal, error) {

	if liquidity.IsNegative() {
		r, err := GetAmount1DeltaWithRoundUp(sqrtRatioAX96, sqrtRatioBX96, liquidity.Neg(), false)
		if err != nil {
			return ZERO, err
		}
		return r.Neg(), nil

	} else {
		r, err := GetAmount1DeltaWithRoundUp(sqrtRatioAX96, sqrtRatioBX96, liquidity, true)
		if err != nil {
			return ZERO, err
		}
		return r, nil
	}
}

func GetAmount0Delta(
	sqrtRatioAX96 decimal.Decimal,
	sqrtRatioBX96 decimal.Decimal,
	liquidity decimal.Decimal,
) (decimal.Decimal, error) {

	if liquidity.IsNegative() {
		r, err := GetAmount0DeltaWithRoundUp(sqrtRatioAX96, sqrtRatioBX96, liquidity.Neg(), false)
		if err != nil {
			return ZERO, err
		}
		return r.Neg(), nil

	} else {
		r, err := GetAmount0DeltaWithRoundUp(sqrtRatioAX96, sqrtRatioBX96, liquidity, true)
		if err != nil {
			return ZERO, err
		}
		return r, nil
	}
}

func GetAmount1DeltaWithRoundUp(
	sqrtRatioAX96 decimal.Decimal,
	sqrtRatioBX96 decimal.Decimal,
	liquidity decimal.Decimal,
	roundUp bool) (decimal.Decimal, error) {
	if sqrtRatioAX96.GreaterThan(sqrtRatioBX96) {
		sqrtRatioAX96 = sqrtRatioBX96
		sqrtRatioBX96 = sqrtRatioAX96
	}
	tmp2, err := MulDivRoundingUp(liquidity, sqrtRatioBX96.Sub(sqrtRatioAX96), Q96)
	if err != nil {
		return ZERO, err
	}
	if roundUp {
		return tmp2, nil
	} else {
		return liquidity.Mul(sqrtRatioBX96.Sub(sqrtRatioAX96)).Div(Q96).RoundDown(0), nil
	}

}
func GetAmount0DeltaWithRoundUp(
	sqrtRatioAX96 decimal.Decimal,
	sqrtRatioBX96 decimal.Decimal,
	liquidity decimal.Decimal,
	roundUp bool) (decimal.Decimal, error) {
	if sqrtRatioAX96.GreaterThan(sqrtRatioBX96) {
		sqrtRatioAX96 = sqrtRatioBX96
		sqrtRatioBX96 = sqrtRatioAX96
	}
	numerator1_bi := liquidity.BigInt()
	numerator1 := decimal.NewFromBigInt(numerator1_bi.Lsh(numerator1_bi, 96), 0)
	numerator2 := sqrtRatioBX96.Sub(sqrtRatioAX96)
	tmp1, err := MulDivRoundingUp(numerator1, numerator2, sqrtRatioBX96)
	if err != nil {
		return ZERO, err
	}
	tmp2, err := MulDivRoundingUp(tmp1, ONE, sqrtRatioAX96)
	if err != nil {
		return ZERO, err
	}
	if roundUp {
		return tmp2, nil
	} else {
		return numerator1.Mul(numerator2).Div(sqrtRatioBX96).RoundDown(0).Div(sqrtRatioAX96).RoundDown(0), nil
	}

}

func SqrtRatioX962HumanPrice(sqrtRatioX96, price *big.Int, decimals0, decimals1 int) float64 {
	squared := new(big.Int).Mul(sqrtRatioX96, sqrtRatioX96)
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals0)), nil)
	squared.Mul(squared, multiplier)
	result := new(big.Int).Mul(squared, price)

	divisor := new(big.Int).Lsh(big.NewInt(1), 192)
	result.Div(result, divisor)

	tenToTheDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals1)), nil)
	quotient, remainder := new(big.Int).QuoRem(result, tenToTheDecimals, new(big.Int))

	intPart, _ := strconv.ParseFloat(quotient.String(), 64)
	remainderPart, _ := strconv.ParseFloat(remainder.String(), 64)
	decimalPart := remainderPart / float64(tenToTheDecimals.Int64())

	return intPart + decimalPart
}

func HumanPrice2SqrtRatioX96(price float64, decimals0, decimals1 int) (*big.Int, error) {
	twoTo192 := new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil)
	twoToDecimals0 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals0)), nil)
	twoToDecimals1 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals1)), nil)

	valueBigFloat := big.NewFloat(price)
	result := new(big.Int)
	valueBigFloat.Mul(valueBigFloat, new(big.Float).SetInt(twoToDecimals1))
	valueBigFloat.Int(result)

	fmt.Printf("%v\n", valueBigFloat.String())
	fmt.Printf("%v\n", result.String())

	numerator := new(big.Int).Mul(result, twoTo192)
	denominator := new(big.Int).Mul(twoToDecimals0, big.NewInt(1))
	divResult := new(big.Int).Div(numerator, denominator)

	return sqrt(divResult)
}

func sqrt(value *big.Int) (*big.Int, error) {
	if value.Sign() < 0 {
		return nil, fmt.Errorf("square root of negative numbers is not supported")
	}

	if value.Cmp(big.NewInt(1)) < 0 {
		return value, nil
	}

	return newtonIteration(value, big.NewInt(1)), nil
}

func newtonIteration(n, x0 *big.Int) *big.Int {
	// x1 = (n / x0 + x0) / 2
	div := new(big.Int).Div(n, x0)
	x1 := new(big.Int).Rsh(div.Add(div, x0), 1)

	if x0.Cmp(x1) == 0 || x0.Cmp(new(big.Int).Sub(x1, big.NewInt(1))) == 0 {
		return x0
	}
	return newtonIteration(n, x1)
}
