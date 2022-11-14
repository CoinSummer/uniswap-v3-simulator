package uniswap_v3_simulator

import (
	"errors"
	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/daoleno/uniswapv3-sdk/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	PoolAddress          string
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
	TickSpacing int64
	Token0      common.Address
	Token1      common.Address
	Fee         FeeAmount
}

func NewPoolConfig(
	TickSpacing int64,
	Token0 common.Address,
	Token1 common.Address,
	Fee FeeAmount,
) *PoolConfig {
	return &PoolConfig{
		TickSpacing: TickSpacing,
		Token0:      Token0,
		Token1:      Token1,
		Fee:         Fee,
	}
}

// core pool
type CorePool struct {
	gorm.Model
	PoolAddress          string `gorm:"index"`
	HasCreated           bool   // has created in db, Flush will set to true
	Token0               string
	Token1               string
	Fee                  FeeAmount
	TickSpacing          int
	MaxLiquidityPerTick  decimal.Decimal
	CurrentBlockNum      int64
	DeployBlockNum       int64
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
		Token0:               snapshot.PoolConfig.Token0.String(),
		Token1:               snapshot.PoolConfig.Token1.String(),
		Fee:                  snapshot.PoolConfig.Fee,
		TickSpacing:          int(snapshot.PoolConfig.TickSpacing),
		MaxLiquidityPerTick:  TickSpacingToMaxLiquidityPerTick(int(snapshot.PoolConfig.TickSpacing)),
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
		Token0:               config.Token0.String(),
		Token1:               config.Token1.String(),
		Fee:                  config.Fee,
		TickSpacing:          int(config.TickSpacing),
		MaxLiquidityPerTick:  TickSpacingToMaxLiquidityPerTick(int(config.TickSpacing)),
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

// 从链上同步数据， 并保存snapshot到数据库(覆盖上一个snapshot)
// 从数据库加载snapshot， 然后检查和最新区块的差距, 并同步到最新区块
// 如果数据库中没有snapshot，则从initialize开始同步所有event
func (p *CorePool) Load() error {
	if p.DeployBlockNum == 0 {
		// todo etherscan api 获取 部署blockNum
	}
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
func (p *CorePool) Burn(owner string, tickLower, tickUpper int, amount decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	position, amount0, amount1, err := p.modifyPosition(owner, tickLower, tickUpper, amount.Neg())
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	amount0 = amount0.Neg()
	amount1 = amount1.Neg()
	if amount0.IsPositive() || amount1.IsPositive() {
		newTokensOwed0 := position.tokensOwed0.Add(amount0)
		newTokensOwed1 := position.tokensOwed1.Add(amount1)
		position.UpdateBurn(newTokensOwed0, newTokensOwed1)
	}
	return amount0, amount1, nil
}

func (p *CorePool) Collect(recipient string, tickLower, tickUpper int, amount0Req, amount1Req decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	err := p.checkTicks(tickLower, tickUpper)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return p.PositionManager.CollectPosition(recipient, tickLower, tickUpper, amount0Req, amount1Req)
}

type swapState struct {
	amountSpecifiedRemaining decimal.Decimal
	amountCalculated         decimal.Decimal
	sqrtPriceX96             decimal.Decimal
	tick                     int
	liquidity                decimal.Decimal
	feeGrowthGlobalX128      decimal.Decimal
}
type StepComputations struct {
	sqrtPriceStartX96 decimal.Decimal
	tickNext          int
	initialized       bool
	sqrtPriceNextX96  decimal.Decimal
	amountIn          decimal.Decimal
	amountOut         decimal.Decimal
	feeAmount         decimal.Decimal
}

func (p *CorePool) handleSwap(zeroForOne bool, amountSpecified decimal.Decimal, optionalSqrtPriceLimitX96 *decimal.Decimal, isStatic bool) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	var sqrtPriceLimitX96 decimal.Decimal
	if optionalSqrtPriceLimitX96 == nil {
		if zeroForOne {
			sqrtPriceLimitX96 = MIN_SQRT_RATIO.Add(decimal.NewFromInt(1))
		} else {
			sqrtPriceLimitX96 = MAX_SQRT_RATIO.Sub(decimal.NewFromInt(1))
		}
	}
	if zeroForOne {
		if !sqrtPriceLimitX96.GreaterThan(MIN_SQRT_RATIO) {
			return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("RATIO_MIN")
		}
		if !sqrtPriceLimitX96.LessThan(p.SqrtPriceX96) {
			return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("RATIO_CURRENT")
		}

	} else {
		if !sqrtPriceLimitX96.LessThan(MAX_SQRT_RATIO) {
			return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("RATIO_MAX")
		}
		if !sqrtPriceLimitX96.GreaterThan(p.SqrtPriceX96) {
			return decimal.Zero, decimal.Zero, decimal.Zero, errors.New("RATIO_CURRENT")
		}
	}
	exactInput := amountSpecified.GreaterThanOrEqual(decimal.Zero)
	state := swapState{
		amountSpecifiedRemaining: amountSpecified,
		amountCalculated:         decimal.Zero,
		sqrtPriceX96:             p.SqrtPriceX96,
		tick:                     p.TickCurrent,
		liquidity:                p.Liquidity,
	}
	if zeroForOne {
		state.feeGrowthGlobalX128 = p.FeeGrowthGlobal0X128
	} else {
		state.feeGrowthGlobalX128 = p.FeeGrowthGlobal1X128
	}
	for {
		// 达到限价或者兑换完成
		if state.amountSpecifiedRemaining.Equal(decimal.Zero) || state.sqrtPriceX96.Equal(sqrtPriceLimitX96) {
			break
		}
		step := StepComputations{
			sqrtPriceStartX96: decimal.Zero, tickNext: 0, initialized: false, sqrtPriceNextX96: decimal.Zero, amountIn: decimal.Zero, amountOut: decimal.Zero, feeAmount: decimal.Zero}
		step.sqrtPriceStartX96 = state.sqrtPriceX96
		tickNext, initialized, err := p.TickManager.GetNextInitializedTick(state.tick, p.TickSpacing, zeroForOne)
		if err != nil {
			return decimal.Zero, decimal.Zero, decimal.Zero, err
		}
		step.tickNext = tickNext
		step.initialized = initialized
		if step.tickNext < MIN_TICK {
			step.tickNext = MIN_TICK
		} else if step.tickNext > MAX_TICK {
			step.tickNext = MAX_TICK
		}
		step.sqrtPriceNextX96, err = GetSqrtRatioAtTick(step.tickNext)
		if err != nil {
			return decimal.Zero, decimal.Zero, decimal.Zero, err
		}
		var sqrtRatioTargetX96 decimal.Decimal
		var b1 bool
		if zeroForOne {
			b1 = step.sqrtPriceNextX96.LessThan(sqrtPriceLimitX96)
		} else {
			b1 = step.sqrtPriceNextX96.GreaterThan(sqrtPriceLimitX96)
		}
		if b1 {
			sqrtRatioTargetX96 = sqrtPriceLimitX96
		} else {
			sqrtRatioTargetX96 = step.sqrtPriceNextX96
		}
		_sqrtPriceX96, _amountIn, _amountOut, _feeAmount, err := utils.ComputeSwapStep(state.sqrtPriceX96.BigInt(), sqrtRatioTargetX96.BigInt(), state.liquidity.BigInt(), state.amountSpecifiedRemaining.BigInt(), constants.FeeAmount(p.Fee))
		state.sqrtPriceX96 = decimal.NewFromBigInt(_sqrtPriceX96, 0)
		step.amountIn = decimal.NewFromBigInt(_amountIn, 0)
		step.amountOut = decimal.NewFromBigInt(_amountOut, 0)
		step.feeAmount = decimal.NewFromBigInt(_feeAmount, 0)
		if err != nil {
			return decimal.Zero, decimal.Zero, decimal.Zero, err
		}
		if exactInput {
			state.amountSpecifiedRemaining = state.amountSpecifiedRemaining.Sub(step.amountIn.Add(step.feeAmount))
			state.amountCalculated = state.amountCalculated.Sub(step.amountOut)
		} else {
			state.amountSpecifiedRemaining = state.amountSpecifiedRemaining.Add(step.amountOut)
			state.amountCalculated = state.amountCalculated.Add(step.amountIn.Add(step.feeAmount))
		}
		if state.liquidity.IsPositive() {
			state.feeGrowthGlobalX128 = state.feeGrowthGlobalX128.Add(step.feeAmount.Mul(Q128).Div(state.liquidity).RoundDown(0))
		}
		if state.sqrtPriceX96.Equal(step.sqrtPriceNextX96) {
			if step.initialized {
				nextTick, err := p.TickManager.GetTickAndInitIfAbsent(step.tickNext)
				if err != nil {
					return decimal.Zero, decimal.Zero, decimal.Zero, err
				}
				var liquidityNet decimal.Decimal
				if isStatic {
					liquidityNet = nextTick.LiquidityNet
				} else {
					if zeroForOne {
						liquidityNet = nextTick.Cross(state.feeGrowthGlobalX128, p.FeeGrowthGlobal1X128)
					} else {
						liquidityNet = nextTick.Cross(p.FeeGrowthGlobal0X128, state.feeGrowthGlobalX128)
					}
				}
				if zeroForOne {
					liquidityNet = liquidityNet.Neg()
				}
				state.liquidity, err = AddDelta(state.liquidity, liquidityNet)
				if err != nil {
					return decimal.Zero, decimal.Zero, decimal.Zero, err
				}

			}
			if zeroForOne {
				state.tick = step.tickNext - 1
			} else {
				state.tick = step.tickNext
			}
		} else if !state.sqrtPriceX96.Equal(step.sqrtPriceStartX96) {
			state.tick, err = GetTickAtSqrtRatio(state.sqrtPriceX96)
			if err != nil {
				return decimal.Zero, decimal.Zero, decimal.Zero, err
			}
		}
	}
	if !isStatic {
		p.SqrtPriceX96 = state.sqrtPriceX96
		if state.tick != p.TickCurrent {
			p.TickCurrent = state.tick
		}
		if !state.liquidity.Equal(p.Liquidity) {
			p.Liquidity = state.liquidity
		}
		if zeroForOne {
			p.FeeGrowthGlobal0X128 = state.feeGrowthGlobalX128
		} else {
			p.FeeGrowthGlobal1X128 = state.feeGrowthGlobalX128
		}
	}
	var amount0, amount1 decimal.Decimal
	if zeroForOne == exactInput {
		amount0 = amountSpecified.Sub(state.amountSpecifiedRemaining)
		amount1 = state.amountCalculated
	} else {
		amount1 = amountSpecified.Sub(state.amountSpecifiedRemaining)
		amount0 = state.amountCalculated
	}
	return amount0, amount1, state.sqrtPriceX96, nil
}

type SwapSolution struct {
	AmountSpecified   decimal.Decimal
	SqrtPriceLimitX96 *decimal.Decimal
}

func (p *CorePool) tryToDryRun(param *UniV3SwapEvent, amountSpec decimal.Decimal, sqrtPriceLimitX96 *decimal.Decimal) bool {
	var zeroForOne = param.Amount0.IsPositive()
	amount0, amount1, priceX96, err := p.handleSwap(zeroForOne, amountSpec, sqrtPriceLimitX96, true)
	if err != nil {
		logrus.Error(err)
		return false
	}
	return amount0.Equal(param.Amount0) && amount1.Equal(param.Amount1) && priceX96.Equal(param.SqrtPriceX96)
}

func incTowardsInfinity(d decimal.Decimal) decimal.Decimal {
	if d.IsZero() {
		logrus.Fatal(d)
	}
	if d.IsPositive() {
		return d.Add(ONE)
	} else {
		return d.Sub(ONE)
	}
}
func (p *CorePool) ResolveInputFromSwapResultEvent(param *UniV3SwapEvent) (decimal.Decimal, *decimal.Decimal, error) {
	solution1 := SwapSolution{SqrtPriceLimitX96: &param.SqrtPriceX96}
	if param.Liquidity.IsZero() {
		solution1.AmountSpecified = incTowardsInfinity(param.Amount0)
	} else {
		solution1.AmountSpecified = param.Amount0
	}

	solution2 := SwapSolution{SqrtPriceLimitX96: &param.SqrtPriceX96}
	if param.Liquidity.IsZero() {
		solution2.AmountSpecified = incTowardsInfinity(param.Amount1)
	} else {
		solution2.AmountSpecified = param.Amount1
	}

	solution3 := SwapSolution{SqrtPriceLimitX96: nil, AmountSpecified: param.Amount0}
	solution4 := SwapSolution{SqrtPriceLimitX96: nil, AmountSpecified: param.Amount1}
	solutionList := []SwapSolution{solution3, solution4}
	if !param.SqrtPriceX96.Equal(p.SqrtPriceX96) {
		if param.Liquidity.Equal(decimal.NewFromInt(-1)) {
			solution5 := SwapSolution{AmountSpecified: param.Amount0, SqrtPriceLimitX96: &param.SqrtPriceX96}
			solution6 := SwapSolution{AmountSpecified: param.Amount1, SqrtPriceLimitX96: &param.SqrtPriceX96}
			solutionList = append(solutionList, solution5)
			solutionList = append(solutionList, solution6)
		}
		solutionList = append(solutionList, solution1)
		solutionList = append(solutionList, solution2)
	}
	for _, solution := range solutionList {
		if p.tryToDryRun(param, solution.AmountSpecified, solution.SqrtPriceLimitX96) {
			return solution.AmountSpecified, solution.SqrtPriceLimitX96, nil
		}
	}
	logrus.Fatal("failed find swap solution")
	return decimal.Zero, nil, nil
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
	return nil
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
		tick, err := p.TickManager.GetTickAndInitIfAbsent(lower)
		if err != nil {
			return nil, err
		}
		flippedLower, err = tick.Update(delta, p.TickCurrent, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128, false, p.MaxLiquidityPerTick)
		if err != nil {
			return nil, err
		}

		tick, err = p.TickManager.GetTickAndInitIfAbsent(upper)
		if err != nil {
			return nil, err
		}
		flippedUpper, err = tick.Update(delta, p.TickCurrent, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128, true, p.MaxLiquidityPerTick)
		if err != nil {
			return nil, err
		}
	}
	fi0, fi1, err := p.TickManager.getFeeGrowthInside(lower, upper, p.TickCurrent, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128)
	if err != nil {
		return nil, err
	}
	err = position.Update(delta, fi0, fi1)
	if err != nil {
		return nil, err
	}
	if delta.IsNegative() {
		if flippedLower {
			p.TickManager.Clear(lower)
		}
		if flippedUpper {
			p.TickManager.Clear(upper)
		}
	}
	return position, nil
}

func (p *CorePool) Flush(db *gorm.DB) error {
	if p.HasCreated {
		return db.Model(p).Updates(p).Error
	} else {
		p.HasCreated = true
		return db.Create(p).Error
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
