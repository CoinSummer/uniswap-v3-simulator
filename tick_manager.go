package uniswap_v3_simulator

import (
	"errors"
	"github.com/shopspring/decimal"
)

type Tick struct {
	TickIndex             int64
	LiquidityGross        decimal.Decimal
	LiquidityNet          decimal.Decimal
	FeeGrowthOutside0X128 decimal.Decimal
	FeeGrowthOutside1X128 decimal.Decimal
}

func (t *Tick) Initialized() bool {
	return !t.LiquidityGross.IsZero()
}

func (t *Tick) Update(
	liquidityDelta decimal.Decimal,
	tickCurrent int64,
	feeGrowthGlobal0X128 decimal.Decimal,
	feeGrowthGlobal1X128 decimal.Decimal,
	upper bool,
	maxLiquidity decimal.Decimal,
) (bool, error) {
	liquidityGrossBefore := t.LiquidityGross
	liquidityGrossAfter, err := LiquidityAddDelta(
		liquidityGrossBefore,
		liquidityDelta,
	)
	if err != nil {
		return false, err
	}
	if liquidityGrossAfter.GreaterThan(maxLiquidity) {
		return false, errors.New("L0")
	}
	flipped := liquidityGrossAfter.IsZero() != liquidityGrossBefore.IsZero()

	if liquidityGrossBefore.IsZero() {
		if t.TickIndex <= tickCurrent {
			t.FeeGrowthOutside0X128 = feeGrowthGlobal0X128
			t.FeeGrowthOutside1X128 = feeGrowthGlobal1X128
		}
	}
	t.LiquidityGross = liquidityGrossAfter
	if upper {
		t.LiquidityNet = t.LiquidityNet.Sub(liquidityDelta)
	} else {
		t.LiquidityNet = t.LiquidityNet.Add(liquidityDelta)
	}
	if t.LiquidityNet.GreaterThan(MaxInt128) {
		return false, OVERFLOW
	}
	if t.LiquidityNet.LessThan(MinInt128) {
		return false, UNDERFLOW
	}
	return flipped, nil
}

func (t *Tick) Cross(
	feeGrowthGlobal0X128 decimal.Decimal,
	feeGrowthGlobal1X128 decimal.Decimal,
) decimal.Decimal {
	t.FeeGrowthOutside0X128 = feeGrowthGlobal0X128.Sub(t.FeeGrowthOutside0X128)
	t.FeeGrowthOutside1X128 = feeGrowthGlobal1X128.Sub(t.FeeGrowthOutside1X128)
	return t.LiquidityNet
}

type TickManager struct {
}
