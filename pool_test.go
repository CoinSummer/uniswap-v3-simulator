package uniswap_v3_simulator

import (
	"fmt"
	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/daoleno/uniswapv3-sdk/utils"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestCorePool_computeSwapStep(t *testing.T) {
	price := utils.EncodeSqrtRatioX96(big.NewInt(1), big.NewInt(1))
	priceTarget := utils.EncodeSqrtRatioX96(big.NewInt(101), big.NewInt(100))
	liquidity := new(big.Int).Mul(big.NewInt(2), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	amount := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	fee := 600
	_, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(price, priceTarget, liquidity, amount, constants.FeeAmount(fee))
	assert.Condition(t, func() bool {
		a, _ := new(big.Int).SetString("9975124224178055", 10)
		return amountIn.Cmp(a) == 0
	}, "returns the correct value for sqrt ratio at min tick")
	assert.Condition(t, func() bool {
		a, _ := new(big.Int).SetString("5988667735148", 10)
		return feeAmount.Cmp(a) == 0
	}, "returns the correct value for sqrt ratio at min tick")
	assert.Condition(t, func() bool {
		a, _ := new(big.Int).SetString("9925619580021728", 10)
		return amountOut.Cmp(a) == 0
	}, "returns the correct value for sqrt ratio at min tick")
	assert.Condition(t, func() bool {
		a := new(big.Int).Add(amountIn, amountOut)
		fmt.Println(a)
		fmt.Println(amount)
		return a.Cmp(amount) < 0
	}, "returns the correct value for sqrt ratio at min tick")

}
