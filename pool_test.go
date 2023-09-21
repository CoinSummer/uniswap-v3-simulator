package uniswap_v3_simulator

import (
	"math/big"
	"testing"

	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/daoleno/uniswapv3-sdk/utils"
	"github.com/stretchr/testify/assert"
)

// All tests from: https://github.com/Uniswap/v3-core/blob/main/test/SwapMath.spec.ts#L20

func TestCorePool_computeSwapStep_ExactAmountInThatGetsCappedAtPriceTargetInOneForZero(t *testing.T) {
	price := utils.EncodeSqrtRatioX96(big.NewInt(1), big.NewInt(1))
	priceTarget := utils.EncodeSqrtRatioX96(big.NewInt(101), big.NewInt(100))
	liquidity := new(big.Int).Mul(big.NewInt(2), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	amount := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	fee := 600
	zeroForOne := false

	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(price, priceTarget, liquidity, amount, constants.FeeAmount(fee))

	assert.Equal(t, amountIn.String(), "9975124224178055")
	assert.Equal(t, feeAmount.String(), "5988667735148")
	assert.Equal(t, amountOut.String(), "9925619580021728")

	assert.Condition(t, func() bool {
		a := new(big.Int).Add(amountIn, amountOut)
		return a.Cmp(amount) < 0
	}, "returns the correct value for sqrt ratio at min tick")

	priceAfterWholeInputAmount, _ := utils.GetNextSqrtPriceFromInput(
		price,
		liquidity,
		amount,
		zeroForOne,
	)

	assert.Equal(t, sqrtQ, priceTarget, "price is capped at price target")
	assert.Condition(t, func() bool {
		return sqrtQ.Cmp(priceAfterWholeInputAmount) < 0
	}, "price is less than price after whole input amount")
}

func TestCorePool_computeSwapStep_ExactAmountOutThatGetsCappedAtPriceTargetInOneForZero(t *testing.T) {
	price := utils.EncodeSqrtRatioX96(big.NewInt(1), big.NewInt(1))
	priceTarget := utils.EncodeSqrtRatioX96(big.NewInt(101), big.NewInt(100))
	liquidity := new(big.Int).Mul(big.NewInt(2), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	amount := new(big.Int).Mul(new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), big.NewInt(-1))
	fee := 600
	zeroForOne := false

	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(price, priceTarget, liquidity, amount, constants.FeeAmount(fee))

	assert.Equal(t, amountIn.String(), "9975124224178055")
	assert.Equal(t, feeAmount.String(), "5988667735148")
	assert.Equal(t, amountOut.String(), "9925619580021728")

	assert.Condition(t, func() bool {
		return amountOut.Cmp(new(big.Int).Mul(amount, big.NewInt(-1))) < 0
	}, "returns the correct value for sqrt ratio at min tick")

	priceAfterWholeInputAmount, _ := utils.GetNextSqrtPriceFromOutput(
		price,
		liquidity,
		new(big.Int).Mul(amount, big.NewInt(-1)),
		zeroForOne,
	)

	assert.Equal(t, sqrtQ, priceTarget, "price is capped at price target")
	assert.Condition(t, func() bool {
		return sqrtQ.Cmp(priceAfterWholeInputAmount) < 0
	}, "price is less than price after whole input amount")
}

func TestCorePool_computeSwapStep_ExactAmountInThatIsFullySpentInOneForZero(t *testing.T) {
	price := utils.EncodeSqrtRatioX96(big.NewInt(1), big.NewInt(1))
	priceTarget := utils.EncodeSqrtRatioX96(big.NewInt(1000), big.NewInt(100))
	liquidity := new(big.Int).Mul(big.NewInt(2), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	amount := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	fee := 600
	zeroForOne := false

	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(price, priceTarget, liquidity, amount, constants.FeeAmount(fee))

	assert.Equal(t, amountIn.String(), "999400000000000000")
	assert.Equal(t, feeAmount.String(), "600000000000000")
	assert.Equal(t, amountOut.String(), "666399946655997866")

	assert.Condition(t, func() bool {
		return amount.Cmp(new(big.Int).Add(amountIn, feeAmount)) == 0
	}, "entire amount is used")

	priceAfterWholeInputAmount, _ := utils.GetNextSqrtPriceFromInput(
		price,
		liquidity,
		new(big.Int).Sub(amount, feeAmount),
		zeroForOne,
	)

	assert.Condition(t, func() bool {
		return sqrtQ.Cmp(priceTarget) < 0
	}, "price does not reach price target")
	assert.Equal(t, sqrtQ, priceAfterWholeInputAmount, "price is equal to price after whole input amount")
}

func TestCorePool_computeSwapStep_ExactAmountOutThatIsFullyReceivedInOneForZero(t *testing.T) {
	price := utils.EncodeSqrtRatioX96(big.NewInt(1), big.NewInt(1))
	priceTarget := utils.EncodeSqrtRatioX96(big.NewInt(10000), big.NewInt(100))
	liquidity := new(big.Int).Mul(big.NewInt(2), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	amount := new(big.Int).Mul(new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), big.NewInt(-1))
	fee := 600
	zeroForOne := false

	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(price, priceTarget, liquidity, amount, constants.FeeAmount(fee))

	assert.Equal(t, amountIn.String(), "2000000000000000000")
	assert.Equal(t, feeAmount.String(), "1200720432259356")

	assert.Condition(t, func() bool {
		return amountOut.Cmp(new(big.Int).Mul(amount, big.NewInt(-1))) == 0
	}, "entire amount is used")

	priceAfterWholeInputAmount, _ := utils.GetNextSqrtPriceFromOutput(
		price,
		liquidity,
		new(big.Int).Mul(amount, big.NewInt(-1)),
		zeroForOne,
	)

	assert.Condition(t, func() bool {
		return sqrtQ.Cmp(priceTarget) < 0
	}, "price does not reach price target")
	assert.Equal(t, sqrtQ, priceAfterWholeInputAmount, "price is equal to price after whole input amount")
}

func TestCorePool_computeSwapStep_AmountOutIsCappedAtTheDesiredAmountOut(t *testing.T) {
	price, _ := new(big.Int).SetString("417332158212080721273783715441582", 10)
	priceTarget, _ := new(big.Int).SetString("1452870262520218020823638996", 10)
	liquidity, _ := new(big.Int).SetString("159344665391607089467575320103", 10)
	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(
		price,
		priceTarget,
		liquidity,
		big.NewInt(-1),
		constants.FeeAmount(1),
	)

	assert.Equal(t, amountIn, big.NewInt(1))
	assert.Equal(t, feeAmount, big.NewInt(1))
	assert.Equal(t, amountOut, big.NewInt(1))
	assert.Equal(t, sqrtQ.String(), "417332158212080721273783715441581")
}

func TestCorePool_computeSwapStep_TargetPriceOf1UsesPartialInputAmount(t *testing.T) {
	amount, _ := new(big.Int).SetString("3915081100057732413702495386755767", 10)
	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(
		big.NewInt(2),
		big.NewInt(1),
		big.NewInt(1),
		amount,
		constants.FeeAmount(1),
	)

	assert.Equal(t, amountIn.String(), "39614081257132168796771975168")
	assert.Equal(t, feeAmount.String(), "39614120871253040049813")
	assert.Condition(t, func() (success bool) {
		a, _ := new(big.Int).SetString("3915081100057732413702495386755767", 10)
		return new(big.Int).Add(amountIn, feeAmount).Cmp(a) <= 0
	})
	assert.Equal(t, amountOut.String(), "0")
	assert.Equal(t, sqrtQ.String(), "1")
}

//     it('entire input amount taken as fee', async () => {
//      const { amountIn, amountOut, sqrtQ, feeAmount } = await swapMath.computeSwapStep(
//        '2413',
//        '79887613182836312',
//        '1985041575832132834610021537970',
//        '10',
//        1872
//      )
//      expect(amountIn).to.eq('0')
//      expect(feeAmount).to.eq('10')
//      expect(amountOut).to.eq('0')
//      expect(sqrtQ).to.eq('2413')
//    })

func TestCorePool_computeSwapStep_EntireInputAmountTakenAsFee(t *testing.T) {
	liquidity, _ := new(big.Int).SetString("1985041575832132834610021537970", 10)
	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(
		big.NewInt(2413),
		big.NewInt(79887613182836312),
		liquidity,
		big.NewInt(10),
		constants.FeeAmount(1872),
	)

	assert.Equal(t, amountIn.String(), "0")
	assert.Equal(t, feeAmount.String(), "10")
	assert.Equal(t, amountOut.String(), "0")
	assert.Equal(t, sqrtQ.String(), "2413")
}

func TestCorePool_computeSwapStep_HandlesIntermediateInsufficientLiquidityInZeroForOneExactOutputCase(t *testing.T) {
	sqrtP, _ := new(big.Int).SetString("20282409603651670423947251286016", 10)
	sqrtPTarget := new(big.Int).Div(new(big.Int).Mul(sqrtP, big.NewInt(11)), big.NewInt(10))

	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(
		sqrtP,
		sqrtPTarget,
		big.NewInt(1024),
		big.NewInt(-4),
		constants.FeeAmount(3000),
	)

	assert.Equal(t, amountOut.String(), "0")
	assert.Equal(t, sqrtQ, sqrtPTarget)
	assert.Equal(t, amountIn.String(), "26215")
	assert.Equal(t, feeAmount.String(), "79")
}

func TestCorePool_computeSwapStep_HandlesIntermediateInsufficientLiquidityInOneForZeroExactOutputCase(t *testing.T) {
	sqrtP, _ := new(big.Int).SetString("20282409603651670423947251286016", 10)
	sqrtPTarget := new(big.Int).Div(new(big.Int).Mul(sqrtP, big.NewInt(9)), big.NewInt(10))

	sqrtQ, amountIn, amountOut, feeAmount, _ := utils.ComputeSwapStep(
		sqrtP,
		sqrtPTarget,
		big.NewInt(1024),
		big.NewInt(-263000),
		constants.FeeAmount(3000),
	)

	assert.Equal(t, amountOut.String(), "26214")
	assert.Equal(t, sqrtQ, sqrtPTarget)
	assert.Equal(t, amountIn.String(), "1")
	assert.Equal(t, feeAmount.String(), "1")
}
