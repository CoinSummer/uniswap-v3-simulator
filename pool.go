package uniswap_v3_simulator

import (
	"errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type FeeAmount int

const (
	FeeAmountLow    FeeAmount = 500
	FeeAmountMedium FeeAmount = 3000
	FeeAmountHigh   FeeAmount = 10000
)

// snapshot
type Snapshot struct {
	Id                   string
	Description          string
	PoolConfig           *PoolConfig
	Token0Balance        decimal.Decimal
	Token1Balance        decimal.Decimal
	SqrtPriceX96         decimal.Decimal
	Liquidity            decimal.Decimal
	TickCurrent          int
	FeeGrowthGlobal0X128 decimal.Decimal
	FeeGrowthGlobal1X128 decimal.Decimal
	TickManager          *TickManager
	PositionManager      *PositionManager
	Timestamp            time.Time
}

// pool config
type PoolConfig struct {
	Id          string
	TickSpacing int
	Token0      string
	Token1      string
	Fee         FeeAmount
}

func NewPoolConfig(
	TickSpacing int,
	Token0 string,
	Token1 string,
	Fee FeeAmount,
) (*PoolConfig, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return &PoolConfig{
		Id:          id.String(),
		TickSpacing: TickSpacing,
		Token0:      Token0,
		Token1:      Token1,
		Fee:         Fee,
	}, nil
}

// core pool
type CorePool struct {
	Token0               string
	Token1               string
	Fee                  FeeAmount
	TickSpacing          int
	MaxLiquidityPerTick  decimal.Decimal
	Token0Balance        decimal.Decimal
	Token1Balance        decimal.Decimal
	SqrtPriceX96         decimal.Decimal
	Liquidity            decimal.Decimal
	TickCurrent          int
	FeeGrowthGlobal0X128 decimal.Decimal
	FeeGrowthGlobal1X128 decimal.Decimal
	TickManager          *TickManager
	PositionManager      *PositionManager
}

func NewCorePoolFromSnapshot(snapshot Snapshot) *CorePool {
	return &CorePool{
		Token0:               snapshot.PoolConfig.Token0,
		Token1:               snapshot.PoolConfig.Token1,
		Fee:                  snapshot.PoolConfig.Fee,
		TickSpacing:          snapshot.PoolConfig.TickSpacing,
		MaxLiquidityPerTick:  TickSpacingToMaxLiquidityPerTick(snapshot.PoolConfig.TickSpacing),
		Token0Balance:        snapshot.Token0Balance,
		Token1Balance:        snapshot.Token1Balance,
		SqrtPriceX96:         snapshot.SqrtPriceX96,
		Liquidity:            snapshot.Liquidity,
		TickCurrent:          snapshot.TickCurrent,
		FeeGrowthGlobal0X128: snapshot.FeeGrowthGlobal0X128,
		FeeGrowthGlobal1X128: snapshot.FeeGrowthGlobal1X128,
		TickManager:          snapshot.TickManager,
		PositionManager:      snapshot.PositionManager,
	}
}
func NewCorePoolFromConfig(config PoolConfig) *CorePool {
	return &CorePool{
		Token0:               config.Token0,
		Token1:               config.Token1,
		Fee:                  config.Fee,
		TickSpacing:          config.TickSpacing,
		MaxLiquidityPerTick:  TickSpacingToMaxLiquidityPerTick(config.TickSpacing),
		Token0Balance:        decimal.Zero,
		Token1Balance:        decimal.Zero,
		SqrtPriceX96:         decimal.Zero,
		Liquidity:            decimal.Zero,
		TickCurrent:          0,
		FeeGrowthGlobal0X128: decimal.Zero,
		FeeGrowthGlobal1X128: decimal.Zero,
		TickManager:          NewTickManager(),
		PositionManager:      NewPositionManager(),
	}
}
func (p *CorePool) Initialize(sqrtPriceX96 decimal.Decimal) error {
	if !p.SqrtPriceX96.IsZero() {
		return errors.New("Already initialized!")
	}
	var err error
	p.TickCurrent, err = GetTickAtSqrtRatio(sqrtPriceX96)
	if err != nil {
		return err
	}
	p.SqrtPriceX96 = sqrtPriceX96
	return nil
}

func (p *CorePool) Mint(recipient string, tickLower, tickUpper int, amount decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	if !amount.GreaterThan(decimal.Zero) {
		return decimal.Zero, decimal.Zero, errors.New("Mint amount should greater than 0")
	}

	_, amount0, amount1, err := p.modifyPosition(recipient, tickLower, tickUpper, amount)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return amount0, amount1, nil
}

func (p *CorePool) checkTicks(tickLower, tickUpper int) error {
	if !(tickLower < tickUpper) {
		return errors.New("tickLower should lower than tickUpper")
	}
	if !(tickLower >= MIN_TICK) {
		return errors.New("tickLower should NOT lower than MIN_TICK")
	}
	if !(tickUpper <= MAX_TICK) {
		return errors.New("tickUpper should NOT greater than MAX_TICK")
	}
}

func (p *CorePool) modifyPosition(owner string, tickLower, tickUpper int, liquidityDelta decimal.Decimal) (*Position, decimal.Decimal, decimal.Decimal, error) {
	err := p.checkTicks(tickLower, tickUpper)
	if err != nil {
		return nil, decimal.Zero, decimal.Zero, err
	}
	amount0 := decimal.Zero
	amount1 := decimal.Zero
	positionView := p.PositionManager.GetPositionReadonly(owner, tickLower, tickUpper)
	if liquidityDelta.IsNegative() {
		negatedLiquidityDelta := liquidityDelta.Neg()
		if !positionView.liquidity.GreaterThanOrEqual(negatedLiquidityDelta) {
			return nil, decimal.Zero, decimal.Zero, errors.New("Liquidity Underflow")
		}
	}
	position, err := p.updatePosition(owner, tickLower, tickUpper, liquidityDelta)
	if err != nil {
		return nil, decimal.Zero, decimal.Zero, err
	}
	if !liquidityDelta.IsZero() {
		if p.TickCurrent < tickLower {
			tmp1, err := GetSqrtRatioAtTick(tickLower)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			tmp2, err := GetSqrtRatioAtTick(tickUpper)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			amount0, err = GetAmount0Delta(tmp1, tmp2, liquidityDelta)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
		} else if p.TickCurrent < tickUpper {
			tmp2, err := GetSqrtRatioAtTick(tickUpper)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			amount0, err = GetAmount0Delta(p.SqrtPriceX96, tmp2, liquidityDelta)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			tmp3, err := GetSqrtRatioAtTick(tickLower)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			amount1, err = GetAmount1Delta(tmp3, p.SqrtPriceX96, liquidityDelta)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			p.Liquidity, err = AddDelta(p.Liquidity, liquidityDelta)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
		} else {
			tmp1, err := GetSqrtRatioAtTick(tickLower)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			tmp2, err := GetSqrtRatioAtTick(tickUpper)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
			amount1, err = GetAmount1Delta(tmp1, tmp2, liquidityDelta)
			if err != nil {
				return nil, decimal.Zero, decimal.Zero, err
			}
		}
	}
	return position, amount0, amount1, nil
}

func (p *CorePool) updatePosition(owner string, lower int, upper int, delta decimal.Decimal) (*Position, error) {
	position := p.PositionManager.GetPositionAndInitIfAbsent(GetPositionKey(owner, lower, upper))
	flippedLower := false
	flippedUpper := false
	if !delta.IsZero() {

	}
}

type ActionType string

const (
	INITIALIZE ActionType = "initialize"
	MINT       ActionType = "mint"
	BURN       ActionType = "burn"
	COLLECT    ActionType = "collect"
	SWAP       ActionType = "swap"
	FORK       ActionType = "fork"
)

type Record struct {
	Id         string
	ActionType ActionType
	Params     interface{}
	Amount0    decimal.Decimal
	Amount1    decimal.Decimal
	Timestamp  time.Time
}
