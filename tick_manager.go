package uniswap_v3_simulator

import (
	"errors"
	"github.com/shopspring/decimal"
	"math"
)

type Tick struct {
	TickIndex             int
	LiquidityGross        decimal.Decimal
	LiquidityNet          decimal.Decimal
	FeeGrowthOutside0X128 decimal.Decimal
	FeeGrowthOutside1X128 decimal.Decimal
}

func NewTick(index int) (*Tick, error) {
	if index > MAX_TICK || index < MIN_TICK {
		return nil, errors.New("TICK")
	} else {
		return &Tick{
			TickIndex:             index,
			LiquidityGross:        decimal.Zero,
			LiquidityNet:          decimal.Zero,
			FeeGrowthOutside0X128: decimal.Zero,
			FeeGrowthOutside1X128: decimal.Zero,
		}, nil
	}
}

func (t *Tick) Initialized() bool {
	return !t.LiquidityGross.IsZero()
}

func (t *Tick) Update(
	liquidityDelta decimal.Decimal,
	tickCurrent int,
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
	ticks map[int]*Tick
}

func NewTickManager() *TickManager {
	return &TickManager{
		ticks: map[int]*Tick{},
	}
}
func (tm *TickManager) GetTickAndInitIfAbsent(index int) (*Tick, error) {
	if tick, ok := tm.ticks[index]; ok {
		return tick, nil
	} else {
		tick, err := NewTick(index)
		if err != nil {
			return nil, err
		}
		tm.ticks[tick.TickIndex] = tick
		return tick, nil
	}
}
func (tm *TickManager) GetTickReadonly(index int) (*Tick, error) {
	if tick, ok := tm.ticks[index]; ok {
		return tick, nil
	} else {
		tick, err := NewTick(index)
		if err != nil {
			return nil, err
		}
		return tick, nil
	}
}
func (tm *TickManager) nextInitializedTick(ticks []*Tick, tick int, lte bool) (*Tick, error) {

	if lte {
		if tm.isBelowSmallest(ticks, tick) {
			return nil, errors.New("BELOW_SMALLEST")
		}
		if tm.isAtOrAboveLargest(ticks, tick) {
			return ticks[len(ticks)-1], nil
		}
		index, err := tm.binarySearch(ticks, tick)
		if err != nil {
			return nil, err
		}
		return ticks[index], nil
	} else {
		if tm.isAtOrAboveLargest(ticks, tick) {
			return nil, errors.New("AT_OR_ABOVE_LARGEST")
		}
		if tm.isBelowSmallest(ticks, tick) {
			return ticks[0], nil
		}
		index, err := tm.binarySearch(ticks, tick)
		if err != nil {
			return nil, err
		}
		return ticks[index+1], nil
	}
}
func (tm *TickManager) GetSortedTicks() []*Tick {
	// todo sort tick
}
func (tm *TickManager) GetNextInitializedTick(tick, tickSpacing int, lte bool) (int, bool, error) {
	sortedTicks := tm.GetSortedTicks()
	compressed := int(math.Floor(float64(tick / tickSpacing)))
	if lte {
		wordPos := compressed >> 8
		minimum := (wordPos << 8) * tickSpacing
		if tm.isBelowSmallest(sortedTicks, tick) {
			return minimum, false, nil
		}
		nextTick, err := tm.nextInitializedTick(sortedTicks, tick, lte)
		if err != nil {
			return 0, false, err
		}
		nextInitializedTick := int(math.Max(float64(minimum), float64(nextTick.TickIndex)))
		return nextInitializedTick, nextInitializedTick == nextTick.TickIndex, nil
	} else {
		wordPos := (compressed + 1) >> 8
		maximum := (((wordPos + 1) << 8) - 1) * tickSpacing
		if tm.isAtOrAboveLargest(sortedTicks, tick) {
			return maximum, false, nil
		}
		nextTick, err := tm.nextInitializedTick(sortedTicks, tick, lte)
		if err != nil {
			return 0, false, err
		}
		nextInitializedTick := int(math.Max(float64(maximum), float64(nextTick.TickIndex)))
		return nextInitializedTick, nextInitializedTick == nextTick.TickIndex, nil
	}
}

func (tm *TickManager) getFeeGrowthInside(tickLower, tickUpper, tickCurrent int, feeGrowthGlobal0X128, feeGrowthGlobal1X128 decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {

}

func (tm *TickManager) isAtOrAboveLargest(sortedTicks []*Tick, tick int) bool {
	if len(sortedTicks) == 0 {
		return false
	}
	return tick > sortedTicks[len(sortedTicks)-1].TickIndex
}

func (tm *TickManager) isBelowSmallest(sortedTicks []*Tick, tick int) bool {
	if len(sortedTicks) == 0 {
		return false
	}
	return tick < sortedTicks[0].TickIndex
}
func (tm *TickManager) binarySearch(sortedTicks []*Tick, tick int) (int, error) {
	if tm.isBelowSmallest(sortedTicks, tick) {
		return 0, errors.New("BELOW_SMALLEST")
	}
	var l = 0
	var r = len(sortedTicks) - 1
	var i = 0
	for {
		i = int(math.Floor(float64((l + r) / 2)))
		if sortedTicks[i].TickIndex <= tick && (i == len(sortedTicks)-1 || sortedTicks[i+1].TickIndex > tick) {
			return i, nil
		}
		if sortedTicks[i].TickIndex < tick {
			l = i + 1
		} else {
			r = i - 1
		}
	}
}
