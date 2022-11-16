package uniswap_v3_simulator

import (
	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestGetSqrtRatioAtTick(t *testing.T) {
	_, err := GetSqrtRatioAtTick(MIN_TICK - 1)
	assert.ErrorIs(t, err, INVALID_TICK, "tick tool small")

	_, err = GetSqrtRatioAtTick(MAX_TICK + 1)
	assert.ErrorIs(t, err, INVALID_TICK, "tick tool large")

	rmax, _ := GetSqrtRatioAtTick(MIN_TICK)
	assert.Equal(t, rmax, MIN_SQRT_RATIO, "returns the correct value for min tick")

	r0, _ := GetSqrtRatioAtTick(0)
	assert.Condition(t, func() (success bool) {
		return r0.Equal(decimal.NewFromBigInt(new(big.Int).Lsh(constants.One, 96), 0))
	})

	rmin, _ := GetSqrtRatioAtTick(MAX_TICK)
	assert.Equal(t, rmin, MAX_SQRT_RATIO, "returns the correct value for max tick")
}

func TestGetTickAtSqrtRatio(t *testing.T) {
	tmin, _ := GetTickAtSqrtRatio(MIN_SQRT_RATIO)
	assert.Equal(t, tmin, MIN_TICK, "returns the correct value for sqrt ratio at min tick")

	tmax, _ := GetTickAtSqrtRatio(MAX_SQRT_RATIO.Sub(ONE))
	assert.Equal(t, tmax, MAX_TICK-1, "returns the correct value for sqrt ratio at max tick")
}
