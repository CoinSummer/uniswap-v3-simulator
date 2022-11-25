package uniswap_v3_simulator

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"sort"
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
			LiquidityGross:        ZERO,
			LiquidityNet:          ZERO,
			FeeGrowthOutside0X128: ZERO,
			FeeGrowthOutside1X128: ZERO,
		}, nil
	}
}

func (t *Tick) Clone() *Tick {
	return &Tick{
		TickIndex:             t.TickIndex,
		LiquidityGross:        t.LiquidityGross,
		LiquidityNet:          t.LiquidityNet,
		FeeGrowthOutside0X128: t.FeeGrowthOutside0X128,
		FeeGrowthOutside1X128: t.FeeGrowthOutside1X128,
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
	Ticks       map[int]*Tick `json:"ticks"`
	SortedTicks []*Tick       `json:"-"`
}

func NewTickManager() *TickManager {
	return &TickManager{
		Ticks: map[int]*Tick{},
	}
}
func (tm *TickManager) Clone() *TickManager {
	var ticks map[int]*Tick
	for k, tick := range tm.Ticks {
		ticks[k] = tick.Clone()
	}
	newM := NewTickManager()
	newM.Ticks = ticks
	newM.SortTicks()
	return newM
}

func (tm *TickManager) GetTickAndInitIfAbsent(index int) (*Tick, error) {

	if tick, ok := tm.Ticks[index]; ok {
		return tick, nil
	} else {
		tick, err := NewTick(index)
		if err != nil {
			return nil, err
		}
		tm.Ticks[tick.TickIndex] = tick
		tm.SortTicks()
		return tick, nil
	}
}
func (tm *TickManager) GetTickReadonly(index int) (*Tick, error) {
	if tick, ok := tm.Ticks[index]; ok {
		return tick.Clone(), nil
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

func (tm *TickManager) SortTicks() {
	tm.SortedTicks = tm.GetSortedTicks()
}
func (tm *TickManager) Clear(tick int) {
	delete(tm.Ticks, tick)
	tm.SortTicks()
}

func (tm *TickManager) GetSortedTicks() []*Tick {
	keys := make([]int, 0, len(tm.Ticks))
	for k, _ := range tm.Ticks {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var result []*Tick
	for _, k := range keys {
		result = append(result, tm.Ticks[k])
	}
	return result
}

func (tm *TickManager) GetNextInitializedTick(tick, tickSpacing int, lte bool) (int, bool, error) {
	sortedTicks := tm.SortedTicks

	compressed := int(math.Floor(float64(tick) / float64(tickSpacing)))

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
		nextInitializedTick := int(math.Min(float64(maximum), float64(nextTick.TickIndex)))
		return nextInitializedTick, nextInitializedTick == nextTick.TickIndex, nil
	}
}

func (tm *TickManager) getFeeGrowthInside(tickLower, tickUpper, tickCurrent int, feeGrowthGlobal0X128, feeGrowthGlobal1X128 decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	_, lok := tm.Ticks[tickLower]
	_, uok := tm.Ticks[tickUpper]
	if !lok || !uok {
		return ZERO, ZERO, errors.New("INVALID_TICK")
	}
	lower, err := tm.GetTickAndInitIfAbsent(tickLower)
	if err != nil {
		return ZERO, ZERO, err
	}
	upper, err := tm.GetTickAndInitIfAbsent(tickUpper)
	if err != nil {
		return ZERO, ZERO, err
	}

	var feeGrowthBelow0X128 decimal.Decimal
	var feeGrowthBelow1X128 decimal.Decimal
	if tickCurrent >= tickLower {
		feeGrowthBelow0X128 = lower.FeeGrowthOutside0X128
		feeGrowthBelow1X128 = lower.FeeGrowthOutside1X128
	} else {
		feeGrowthBelow0X128 = feeGrowthGlobal0X128.Sub(lower.FeeGrowthOutside0X128)
		feeGrowthBelow1X128 = feeGrowthGlobal1X128.Sub(lower.FeeGrowthOutside1X128)
	}
	var feeGrowthAbove0X128 decimal.Decimal
	var feeGrowthAbove1X128 decimal.Decimal
	if tickCurrent < tickUpper {
		feeGrowthAbove0X128 = upper.FeeGrowthOutside0X128
		feeGrowthAbove1X128 = upper.FeeGrowthOutside1X128
	} else {
		feeGrowthAbove0X128 = feeGrowthGlobal0X128.Sub(upper.FeeGrowthOutside0X128)
		feeGrowthAbove1X128 = feeGrowthGlobal1X128.Sub(upper.FeeGrowthOutside1X128)
	}

	result1, err := Mod256Sub(feeGrowthGlobal0X128, feeGrowthBelow0X128)
	if err != nil {
		return ZERO, ZERO, err
	}
	result1 = result1.Sub(feeGrowthAbove0X128)
	result2, err := Mod256Sub(feeGrowthGlobal1X128, feeGrowthBelow1X128)
	if err != nil {
		return ZERO, ZERO, err
	}
	result2 = result2.Sub(feeGrowthAbove1X128)
	return result1, result2, nil
}

func (tm *TickManager) isAtOrAboveLargest(sortedTicks []*Tick, tick int) bool {
	if len(sortedTicks) == 0 {
		return false
	}
	return tick >= sortedTicks[len(sortedTicks)-1].TickIndex
}

func (tm *TickManager) isBelowSmallest(sortedTicks []*Tick, tick int) bool {
	if len(sortedTicks) == 0 {
		return false
	}
	return tick < sortedTicks[0].TickIndex
}

func (tm *TickManager) binarySearch(ticks []*Tick, tick int) (int, error) {
	if tm.isBelowSmallest(ticks, tick) {
		return 0, errors.New("tick is below smallest tick")
	}

	l := 0
	r := len(ticks) - 1
	i := 0
	for {
		i = int(math.Floor(float64(l+r) / 2))
		if ticks[i].TickIndex <= tick && (i == len(ticks)-1 || ticks[i+1].TickIndex > tick) {
			return i, nil
		}
		if ticks[i].TickIndex < tick {
			l = i + 1
		} else {
			r = i - 1
		}
	}
}

func (nc *TickManager) GormDataType() string {
	return "LONGTEXT"
}

func (j *TickManager) Scan(value interface{}) error {
	var err error
	switch v := value.(type) {
	case []byte:
		{
			err = json.Unmarshal(v, j)
		}
	case string:
		{
			err = json.Unmarshal([]byte(v), j)
		}
	case nil:
		return nil
	default:
		err = errors.New(fmt.Sprint("Failed to unmarshal TickManager value:", value))
	}
	if err == nil {
		j.SortTicks()
	}
	return err
}

func (j *TickManager) Value() (driver.Value, error) {
	bs, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}
