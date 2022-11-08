package uniswap_v3_simulator

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

type Position struct {
	liquidity                decimal.Decimal
	feeGrowthInside0LastX128 decimal.Decimal
	feeGrowthInside1LastX128 decimal.Decimal
	tokensOwed0              decimal.Decimal
	tokensOwed1              decimal.Decimal
}

func NewPosition() *Position {
	return &Position{
		liquidity:                decimal.Zero,
		feeGrowthInside0LastX128: decimal.Zero,
		feeGrowthInside1LastX128: decimal.Zero,
		tokensOwed0:              decimal.Zero,
		tokensOwed1:              decimal.Zero,
	}

}
func (p *Position) Clone() *Position {
	return &Position{
		liquidity:                p.liquidity,
		feeGrowthInside0LastX128: p.feeGrowthInside0LastX128,
		feeGrowthInside1LastX128: p.feeGrowthInside1LastX128,
		tokensOwed0:              p.tokensOwed0,
		tokensOwed1:              p.tokensOwed1,
	}
}
func (p *Position) Update(
	liquidityDelta decimal.Decimal,
	feeGrowthInside0X128 decimal.Decimal,
	feeGrowthInside1X128 decimal.Decimal,
) error {
	var liquidityNext decimal.Decimal
	var err error
	if liquidityDelta.IsZero() {
		if p.liquidity.LessThanOrEqual(decimal.Zero) {
			return errors.New("NP")
		}
		liquidityNext = p.liquidity
	} else {
		liquidityNext, err = LiquidityAddDelta(p.liquidity, liquidityDelta)
		if err != nil {
			return err
		}
	}
	tokensOwed0 := feeGrowthInside0X128.Sub(p.feeGrowthInside0LastX128).Mul(p.liquidity).Div(Q128)
	tokensOwed1 := feeGrowthInside1X128.Sub(p.feeGrowthInside1LastX128).Mul(p.liquidity).Div(Q128)
	if !liquidityDelta.IsZero() {
		p.liquidity = liquidityNext
	}
	p.feeGrowthInside0LastX128 = feeGrowthInside0X128
	p.feeGrowthInside1LastX128 = feeGrowthInside1X128

	if tokensOwed0.GreaterThan(decimal.Zero) || tokensOwed1.GreaterThan(decimal.Zero) {
		p.tokensOwed0 = p.tokensOwed0.Add(tokensOwed0)
		p.tokensOwed1 = p.tokensOwed1.Add(tokensOwed1)
	}
	return nil
}
func (p *Position) UpdateBurn(
	newTokensOwed0 decimal.Decimal,
	newTokensOwed1 decimal.Decimal,
) {
	p.tokensOwed0 = newTokensOwed0
	p.tokensOwed1 = newTokensOwed1
}
func (p *Position) IsEmpty() bool {
	return p.liquidity.IsZero() && p.tokensOwed0.IsZero() && p.tokensOwed1.IsZero()
}

func GetPositionKey(owner string, tickLower int64, tickUpper int64) string {
	return fmt.Sprintf("%s_%d_%d", owner, tickLower, tickUpper)
}

type PositionManager struct {
	positions map[string]*Position
}

func NewPositionManager() *PositionManager {
	return &PositionManager{
		positions: map[string]*Position{},
	}
}

func (pm *PositionManager) Set(key string, position *Position) {
	pm.positions[key] = position
}
func (pm *PositionManager) Clear(key string) {
	delete(pm.positions, key)
}
func (pm *PositionManager) GetPositionAndInitIfAbsent(key string) *Position {
	if v, ok := pm.positions[key]; ok {
		return v
	}
	newP := &Position{}
	pm.Set(key, newP)
	return newP
}
func (pm *PositionManager) GetPositionReadonly(owner string, tickLower int64, tickUpper int64) *Position {
	key := GetPositionKey(owner, tickLower, tickUpper)
	if v, ok := pm.positions[key]; ok {
		return v.Clone()
	}
	return NewPosition()
}
func (pm *PositionManager) CollectPosition(owner string, tickLower int64, tickUpper int64, amount0Requested, amount1Requested decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	if amount0Requested.LessThan(decimal.Zero) || amount1Requested.LessThan(decimal.Zero) {
		return decimal.Zero, decimal.Zero, errors.New("amounts requested should be positive")
	}
	key := GetPositionKey(owner, tickLower, tickUpper)
	if v, ok := pm.positions[key]; ok {
		positionToCollect := v
		var amount0 decimal.Decimal
		if amount0Requested.GreaterThan(positionToCollect.tokensOwed0) {
			amount0 = positionToCollect.tokensOwed0
		} else {
			amount0 = amount0Requested
		}
		var amount1 decimal.Decimal
		if amount1Requested.GreaterThan(positionToCollect.tokensOwed1) {
			amount1 = positionToCollect.tokensOwed1
		} else {
			amount1 = amount1Requested
		}
		if amount0.GreaterThan(decimal.Zero) || amount1.GreaterThan(decimal.Zero) {
			positionToCollect.UpdateBurn(positionToCollect.tokensOwed0.Sub(amount0), positionToCollect.tokensOwed1.Sub(amount1))
		}
		if positionToCollect.IsEmpty() {
			pm.Clear(key)
		}
		return amount0, amount1, nil
	} else {
		return decimal.Zero, decimal.Zero, nil
	}

}
