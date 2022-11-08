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
	TickCurrent          int64
	FeeGrowthGlobal0X128 decimal.Decimal
	FeeGrowthGlobal1X128 decimal.Decimal
	TickManager          *TickManager
	PositionManager      *PositionManager
	Timestamp            time.Time
}

// pool config
type PoolConfig struct {
	Id          string
	TickSpacing int64
	Token0      string
	Token1      string
	Fee         FeeAmount
}

func NewPoolConfig(
	TickSpacing int64,
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
	TickSpacing          int64
	MaxLiquidityPerTick  decimal.Decimal
	Token0Balance        decimal.Decimal
	Token1Balance        decimal.Decimal
	SqrtPriceX96         decimal.Decimal
	Liquidity            decimal.Decimal
	TickCurrent          int64
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
	//p.TickCurrent = Get
	return nil
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
