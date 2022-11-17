package uniswap_v3_simulator

//
//import (
//	"errors"
//	"github.com/daoleno/uniswapv3-sdk/constants"
//	"github.com/daoleno/uniswapv3-sdk/utils"
//	"github.com/shopspring/decimal"
//	"math/big"
//)
//
//func ComputeSwapStep(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, amountRemaining decimal.Decimal, feePips FeeAmount) (sqrtRatioNextX96, amountIn, amountOut, feeAmount decimal.Decimal, err error) {
//
//	zeroForOne := sqrtRatioCurrentX96.GreaterThanOrEqual(sqrtRatioTargetX96)
//	exactIn := !amountRemaining.IsNegative()
//	if exactIn {
//		amountRemainingLessFee := amountRemaining.Mul(MAX_FEE.Sub(decimal.NewFromInt(int64(feePips)))).Div(MAX_FEE).RoundDown(0)
//		if zeroForOne {
//			amountIn, err = GetAmount0DeltaWithRoundUp(sqrtRatioTargetX96, sqrtRatioCurrentX96, liquidity, true)
//			if err != nil {
//				return
//			}
//			amountIn, err = GetAmount1DeltaWithRoundUp(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, true)
//			if err != nil {
//				return
//			}
//		}
//		if amountRemainingLessFee.GreaterThanOrEqual(amountIn) {
//			sqrtRatioNextX96 = sqrtRatioTargetX96
//		} else {
//			sqrtRatioNextX96, err = GetNextSqrtPriceFromInput(sqrtRatioCurrentX96, liquidity, amountRemainingLessFee, zeroForOne)
//			if err != nil {
//				return
//			}
//		}
//	} else {
//		if zeroForOne {
//			amountOut, err = GetAmount1DeltaWithRoundUp(sqrtRatioTargetX96, sqrtRatioCurrentX96, liquidity, false)
//			if err != nil {
//				return
//			}
//			amountOut, err = GetAmount0DeltaWithRoundUp(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, false)
//			if err != nil {
//				return
//			}
//		}
//		if amountRemaining.Neg().GreaterThanOrEqual(amountOut) {
//
//			sqrtRatioNextX96 = sqrtRatioTargetX96
//		} else {
//			sqrtRatioNextX96, err = GetNextSqrtPriceFromOutput(sqrtRatioCurrentX96, liquidity, amountRemaining.Neg(), zeroForOne)
//			if err != nil {
//				return
//			}
//
//		}
//	}
//	max := sqrtRatioTargetX96.Equal(sqrtRatioNextX96)
//	if zeroForOne {
//		if !(max && exactIn) {
//			amountIn, err = GetAmount0DeltaWithRoundUp(sqrtRatioNextX96, sqrtRatioCurrentX96, liquidity, true)
//			if err != nil {
//				return
//			}
//		}
//		if !(max && !exactIn) {
//			amountOut, err = GetAmount1DeltaWithRoundUp(sqrtRatioNextX96, sqrtRatioCurrentX96, liquidity, false)
//			if err != nil {
//				return
//			}
//		}
//	} else {
//		if !(max && exactIn) {
//			amountIn, err = GetAmount1DeltaWithRoundUp(sqrtRatioCurrentX96, sqrtRatioNextX96, liquidity, true)
//			if err != nil {
//				return
//			}
//		}
//		if !(max && !exactIn) {
//			amountOut, err = GetAmount0DeltaWithRoundUp(sqrtRatioCurrentX96, sqrtRatioNextX96, liquidity, false)
//			if err != nil {
//				return
//			}
//		}
//	}
//	if !exactIn && amountOut.GreaterThan(amountRemaining.Neg()) {
//		amountOut = amountRemaining.Neg()
//	}
//	if exactIn && sqrtRatioNextX96.Equal(sqrtRatioTargetX96) {
//		feeAmount = amountRemaining.Sub(amountIn)
//	} else {
//		feeAmount, err = MulDivRoundingUp(amountIn, decimal.NewFromInt(int64(feePips)), MAX_FEE.Sub(decimal.NewFromInt(int64(feePips))))
//		if err != nil {
//			return
//		}
//	}
//	return
//}
//
//func GetNextSqrtPriceFromInput(
//	sqrtPX96,
//	liquidity,
//	amountIn decimal.Decimal,
//	zeroForOne bool) (decimal.Decimal, error) {
//	if !sqrtPX96.IsPositive() || !liquidity.IsPositive() {
//		return decimal.Zero, errors.New("sqrtPx96 and liquidity must be positive")
//	}
//	if zeroForOne {
//		utils.GetNextSqrtPriceFromOutput()
//	}
//}
//
//func getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amount decimal.Decimal, add bool) (decimal.Decimal, error) {
//	if amount.IsZero() {
//		return sqrtPX96, nil
//	}
//
//	numerator1 := new(big.Int).Lsh(liquidity.BigInt(), 96)
//	if add {
//		product := multiplyIn256(amount, sqrtPX96)
//		if new(big.Int).Div(product, amount).Cmp(sqrtPX96) == 0 {
//			denominator := addIn256(numerator1, product)
//			if denominator.Cmp(numerator1) >= 0 {
//				return MulDivRoundingUp(numerator1, sqrtPX96, denominator), nil
//			}
//		}
//		return MulDivRoundingUp(numerator1, constants.One, new(big.Int).Add(new(big.Int).Div(numerator1, sqrtPX96), amount)), nil
//	} else {
//		product := multiplyIn256(amount, sqrtPX96)
//		if new(big.Int).Div(product, amount).Cmp(sqrtPX96) != 0 {
//			return nil, ErrInvariant
//		}
//		if numerator1.Cmp(product) <= 0 {
//			return nil, ErrInvariant
//		}
//		denominator := new(big.Int).Sub(numerator1, product)
//		return MulDivRoundingUp(numerator1, sqrtPX96, denominator), nil
//	}
//}
//
//func getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amount *big.Int, add bool) (*big.Int, error) {
//	if add {
//		var quotient *big.Int
//		if amount.Cmp(MaxUint160) <= 0 {
//			quotient = new(big.Int).Div(new(big.Int).Lsh(amount, 96), liquidity)
//		} else {
//			quotient = new(big.Int).Div(new(big.Int).Mul(amount, constants.Q96), liquidity)
//		}
//		return new(big.Int).Add(sqrtPX96, quotient), nil
//	}
//
//	quotient := MulDivRoundingUp(amount, constants.Q96, liquidity)
//	if sqrtPX96.Cmp(quotient) <= 0 {
//		return nil, ErrInvariant
//	}
//	return new(big.Int).Sub(sqrtPX96, quotient), nil
//}
