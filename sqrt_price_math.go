package uniswap_v3_simulator

import "github.com/shopspring/decimal"

func GetAmount1Delta(
	sqrtRatioAX96 decimal.Decimal,
	sqrtRatioBX96 decimal.Decimal,
	liquidity decimal.Decimal,
) (decimal.Decimal, error) {

	if liquidity.IsNegative() {
		r, err := GetAmount1DeltaWithRoundUp(sqrtRatioAX96, sqrtRatioBX96, liquidity.Neg(), false)
		if err != nil {
			return decimal.Zero, err
		}
		return r, nil

	} else {
		r, err := GetAmount1DeltaWithRoundUp(sqrtRatioAX96, sqrtRatioBX96, liquidity, true)
		if err != nil {
			return decimal.Zero, err
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
			return decimal.Zero, err
		}
		return r, nil

	} else {
		r, err := GetAmount0DeltaWithRoundUp(sqrtRatioAX96, sqrtRatioBX96, liquidity, true)
		if err != nil {
			return decimal.Zero, err
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
	numerator1_bi := liquidity.BigInt()
	numerator1 := decimal.NewFromBigInt(numerator1_bi.Lsh(numerator1_bi, 96), 0)
	numerator2 := sqrtRatioBX96.Sub(sqrtRatioAX96)

	tmp1, err := MulDivRoundingUp(numerator1, numerator2, sqrtRatioBX96)
	if err != nil {
		return decimal.Zero, err
	}
	tmp2, err := MulDivRoundingUp(liquidity, sqrtRatioBX96.Sub(sqrtRatioAX96), Q96)
	if err != nil {
		return decimal.Zero, err
	}
	if roundUp {
		return tmp2, nil
	} else {
		return liquidity.Mul(sqrtRatioBX96.Sub(sqrtRatioAX96)).Div(Q96).Floor(), nil
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
		return decimal.Zero, err
	}
	tmp2, err := MulDivRoundingUp(tmp1, decimal.NewFromInt(1), sqrtRatioAX96)
	if err != nil {
		return decimal.Zero, err
	}
	if roundUp {
		return tmp2, nil
	} else {
		return numerator1.Mul(numerator2).Div(sqrtRatioBX96).Floor().Div(sqrtRatioAX96).Floor(), nil
	}

}
