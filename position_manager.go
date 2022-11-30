package uniswap_v3_simulator

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

type Position struct {
	Liquidity                decimal.Decimal
	FeeGrowthInside0LastX128 decimal.Decimal
	FeeGrowthInside1LastX128 decimal.Decimal
	TokensOwed0              decimal.Decimal
	TokensOwed1              decimal.Decimal
}

func NewPosition() *Position {
	return &Position{
		Liquidity:                ZERO,
		FeeGrowthInside0LastX128: ZERO,
		FeeGrowthInside1LastX128: ZERO,
		TokensOwed0:              ZERO,
		TokensOwed1:              ZERO,
	}
}
func (p *Position) Clone() *Position {
	return &Position{
		Liquidity:                p.Liquidity,
		FeeGrowthInside0LastX128: p.FeeGrowthInside0LastX128,
		FeeGrowthInside1LastX128: p.FeeGrowthInside1LastX128,
		TokensOwed0:              p.TokensOwed0,
		TokensOwed1:              p.TokensOwed1,
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
		if p.Liquidity.LessThanOrEqual(ZERO) {
			return errors.New("NP")
		}
		liquidityNext = p.Liquidity
	} else {
		liquidityNext, err = LiquidityAddDelta(p.Liquidity, liquidityDelta)
		if err != nil {
			return err
		}
	}
	tokensOwed0 := feeGrowthInside0X128.Sub(p.FeeGrowthInside0LastX128).Mul(p.Liquidity).Div(Q128).RoundDown(0)
	tokensOwed1 := feeGrowthInside1X128.Sub(p.FeeGrowthInside1LastX128).Mul(p.Liquidity).Div(Q128).RoundDown(0)
	if !liquidityDelta.IsZero() {
		p.Liquidity = liquidityNext
	}
	p.FeeGrowthInside0LastX128 = feeGrowthInside0X128
	p.FeeGrowthInside1LastX128 = feeGrowthInside1X128

	if tokensOwed0.GreaterThan(ZERO) || tokensOwed1.GreaterThan(ZERO) {
		p.TokensOwed0 = p.TokensOwed0.Add(tokensOwed0)
		p.TokensOwed1 = p.TokensOwed1.Add(tokensOwed1)
	}
	return nil
}
func (p *Position) UpdateBurn(
	newTokensOwed0 decimal.Decimal,
	newTokensOwed1 decimal.Decimal,
) {
	p.TokensOwed0 = newTokensOwed0
	p.TokensOwed1 = newTokensOwed1
}
func (p *Position) IsEmpty() bool {
	return p.Liquidity.IsZero() && p.TokensOwed0.IsZero() && p.TokensOwed1.IsZero()
}

func GetPositionKey(owner string, tickLower int, tickUpper int) string {
	return fmt.Sprintf("%s_%d_%d", owner, tickLower, tickUpper)
}

type PositionManager struct {
	Positions map[string]*Position
}

func NewPositionManager() *PositionManager {
	return &PositionManager{
		Positions: map[string]*Position{},
	}
}

func (pm *PositionManager) Clone() *PositionManager {
	newP := NewPositionManager()
	ps := make(map[string]*Position, len(pm.Positions))
	for s, position := range pm.Positions {
		ps[s] = position.Clone()
	}
	newP.Positions = ps
	return newP
}
func (pm *PositionManager) Set(key string, position *Position) {
	pm.Positions[key] = position
}
func (pm *PositionManager) Clear(key string) {
	delete(pm.Positions, key)
}
func (pm *PositionManager) GetPositionAndInitIfAbsent(key string) *Position {
	if v, ok := pm.Positions[key]; ok {
		return v
	}
	newP := NewPosition()
	pm.Set(key, newP)
	return newP
}
func (pm *PositionManager) GetPositionReadonly(owner string, tickLower int, tickUpper int) *Position {
	key := GetPositionKey(owner, tickLower, tickUpper)
	if v, ok := pm.Positions[key]; ok {
		// todo : clone? or not.
		return v.Clone()
	}
	return NewPosition()
}
func (pm *PositionManager) CollectPosition(owner string, tickLower int, tickUpper int, amount0Requested, amount1Requested decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	if amount0Requested.LessThan(ZERO) || amount1Requested.LessThan(ZERO) {
		return ZERO, ZERO, errors.New("amounts requested should be positive")
	}
	key := GetPositionKey(owner, tickLower, tickUpper)
	if v, ok := pm.Positions[key]; ok {
		positionToCollect := v
		var amount0 decimal.Decimal
		if amount0Requested.GreaterThan(positionToCollect.TokensOwed0) {
			amount0 = positionToCollect.TokensOwed0
		} else {
			amount0 = amount0Requested
		}
		var amount1 decimal.Decimal
		if amount1Requested.GreaterThan(positionToCollect.TokensOwed1) {
			amount1 = positionToCollect.TokensOwed1
		} else {
			amount1 = amount1Requested
		}
		if amount0.GreaterThan(ZERO) || amount1.GreaterThan(ZERO) {
			positionToCollect.UpdateBurn(positionToCollect.TokensOwed0.Sub(amount0), positionToCollect.TokensOwed1.Sub(amount1))
		}
		if positionToCollect.IsEmpty() {
			pm.Clear(key)
		}
		return amount0, amount1, nil
	} else {
		return ZERO, ZERO, nil
	}

}
func (nc *PositionManager) GormDataType() string {
	return "LONGTEXT"
}

func (j *PositionManager) Scan(value interface{}) error {
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
	return err
}

func (j *PositionManager) Value() (driver.Value, error) {
	bs, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}
