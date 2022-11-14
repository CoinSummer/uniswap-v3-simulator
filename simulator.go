package uniswap_v3_simulator

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	blockingester "gitlab.com/CoinSummer/Base/block-ingester"
	"gorm.io/gorm"
	"math/big"
	"strings"
)

var (
	ABI              = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount","type":"uint128"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"}],"name":"Burn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"address","name":"recipient","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount0","type":"uint128"},{"indexed":false,"internalType":"uint128","name":"amount1","type":"uint128"}],"name":"Collect","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint128","name":"amount0","type":"uint128"},{"indexed":false,"internalType":"uint128","name":"amount1","type":"uint128"}],"name":"CollectProtocol","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"paid0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"paid1","type":"uint256"}],"name":"Flash","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint16","name":"observationCardinalityNextOld","type":"uint16"},{"indexed":false,"internalType":"uint16","name":"observationCardinalityNextNew","type":"uint16"}],"name":"IncreaseObservationCardinalityNext","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"indexed":false,"internalType":"int24","name":"tick","type":"int24"}],"name":"Initialize","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount","type":"uint128"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"}],"name":"Mint","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint8","name":"feeProtocol0Old","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol1Old","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol0New","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol1New","type":"uint8"}],"name":"SetFeeProtocol","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"int256","name":"amount0","type":"int256"},{"indexed":false,"internalType":"int256","name":"amount1","type":"int256"},{"indexed":false,"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"indexed":false,"internalType":"uint128","name":"liquidity","type":"uint128"},{"indexed":false,"internalType":"int24","name":"tick","type":"int24"}],"name":"Swap","type":"event"},{"inputs":[{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount","type":"uint128"}],"name":"burn","outputs":[{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount0Requested","type":"uint128"},{"internalType":"uint128","name":"amount1Requested","type":"uint128"}],"name":"collect","outputs":[{"internalType":"uint128","name":"amount0","type":"uint128"},{"internalType":"uint128","name":"amount1","type":"uint128"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint128","name":"amount0Requested","type":"uint128"},{"internalType":"uint128","name":"amount1Requested","type":"uint128"}],"name":"collectProtocol","outputs":[{"internalType":"uint128","name":"amount0","type":"uint128"},{"internalType":"uint128","name":"amount1","type":"uint128"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"factory","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"fee","outputs":[{"internalType":"uint24","name":"","type":"uint24"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"feeGrowthGlobal0X128","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"feeGrowthGlobal1X128","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"flash","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint16","name":"observationCardinalityNext","type":"uint16"}],"name":"increaseObservationCardinalityNext","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"liquidity","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"maxLiquidityPerTick","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount","type":"uint128"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"mint","outputs":[{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"observations","outputs":[{"internalType":"uint32","name":"blockTimestamp","type":"uint32"},{"internalType":"int56","name":"tickCumulative","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityCumulativeX128","type":"uint160"},{"internalType":"bool","name":"initialized","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint32[]","name":"secondsAgos","type":"uint32[]"}],"name":"observe","outputs":[{"internalType":"int56[]","name":"tickCumulatives","type":"int56[]"},{"internalType":"uint160[]","name":"secondsPerLiquidityCumulativeX128s","type":"uint160[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"Positions","outputs":[{"internalType":"uint128","name":"liquidity","type":"uint128"},{"internalType":"uint256","name":"feeGrowthInside0LastX128","type":"uint256"},{"internalType":"uint256","name":"feeGrowthInside1LastX128","type":"uint256"},{"internalType":"uint128","name":"tokensOwed0","type":"uint128"},{"internalType":"uint128","name":"tokensOwed1","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"protocolFees","outputs":[{"internalType":"uint128","name":"token0","type":"uint128"},{"internalType":"uint128","name":"token1","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint8","name":"feeProtocol0","type":"uint8"},{"internalType":"uint8","name":"feeProtocol1","type":"uint8"}],"name":"setFeeProtocol","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"slot0","outputs":[{"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"internalType":"int24","name":"tick","type":"int24"},{"internalType":"uint16","name":"observationIndex","type":"uint16"},{"internalType":"uint16","name":"observationCardinality","type":"uint16"},{"internalType":"uint16","name":"observationCardinalityNext","type":"uint16"},{"internalType":"uint8","name":"feeProtocol","type":"uint8"},{"internalType":"bool","name":"unlocked","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"}],"name":"snapshotCumulativesInside","outputs":[{"internalType":"int56","name":"tickCumulativeInside","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityInsideX128","type":"uint160"},{"internalType":"uint32","name":"secondsInside","type":"uint32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"bool","name":"zeroForOne","type":"bool"},{"internalType":"int256","name":"amountSpecified","type":"int256"},{"internalType":"uint160","name":"sqrtPriceLimitX96","type":"uint160"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"swap","outputs":[{"internalType":"int256","name":"amount0","type":"int256"},{"internalType":"int256","name":"amount1","type":"int256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"int16","name":"","type":"int16"}],"name":"tickBitmap","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"tickSpacing","outputs":[{"internalType":"int24","name":"","type":"int24"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"int24","name":"","type":"int24"}],"name":"Ticks","outputs":[{"internalType":"uint128","name":"liquidityGross","type":"uint128"},{"internalType":"int128","name":"liquidityNet","type":"int128"},{"internalType":"uint256","name":"feeGrowthOutside0X128","type":"uint256"},{"internalType":"uint256","name":"feeGrowthOutside1X128","type":"uint256"},{"internalType":"int56","name":"tickCumulativeOutside","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityOutsideX128","type":"uint160"},{"internalType":"uint32","name":"secondsOutside","type":"uint32"},{"internalType":"bool","name":"initialized","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`
	TOPIC_INITIALIZE = common.HexToHash("0x98636036cb66a9c19a37435efc1e90142190214e8abeb821bdba3f2990dd4c95")
	TOPIC_BURN       = common.HexToHash("0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c")
	TOPIC_SWAP       = common.HexToHash("0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67")
	TOPIC_MINT       = common.HexToHash("0x7a53080ba414158be7ec69b987b5fb7d07dee101fe85488f0853ae16239d0bde")
)

type Simulator struct {
	pools        map[common.Address]*CorePool
	Abi          abi.ABI
	InitializeID common.Hash
	MintID       common.Hash
	BurnID       common.Hash
	SwapID       common.Hash
	ingestor     *blockingester.BlockIngester
	rpc          *blockingester.EthRpcClientPool
	wss          *ethclient.Client
	db           *gorm.DB
	ctx          context.Context
}

// SYNC from univ3 created
func NewPoolManager(startBlock int64, dbFile string, wss string, rpcs []string) *Simulator {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	rpc, err := blockingester.NewClientFactory(rpcs)
	if err != nil {
		panic(err)
	}
	wssClient, err := ethclient.Dial(wss)
	if err != nil {
		panic(err)
	}
	pm := &Simulator{
		pools: map[common.Address]*CorePool{},
		rpc:   rpc,
		wss:   wssClient,
		db:    db,
		ctx:   context.Background(),
	}
	a, err := abi.JSON(strings.NewReader(ABI))
	if err != nil {
		logrus.Fatal(err)
	}
	pm.Abi = a
	pm.InitializeID = a.Events["Initialize"].ID
	logrus.Infof(pm.InitializeID.String())
	pm.MintID = a.Events["Mint"].ID
	pm.BurnID = a.Events["Burn"].ID
	pm.SwapID = a.Events["Swap"].ID
	ingester := blockingester.LoadOrCreateBlockIngester("arbitary", db, wssClient, big.NewInt(startBlock), true, true, pm, context.Background())
	pm.ingestor = ingester
	return pm
}

// get pool config from node
func (pm *Simulator) GetPoolConfig() (*PoolConfig, error) {
	return nil, nil
}

func (pm *Simulator) InitPool(log *types.Log) (*CorePool, error) {
	if _, exist := pm.pools[log.Address]; exist {
		return nil, fmt.Errorf("pool exists %s", log.Address)
	}

	initialze, err := parseUniv3InitializeEvent(log)
	if err != nil {
		return nil, err
	}

	logrus.Infof("initialize pool: %s, block: %d tx: %s", log.Address, log.BlockNumber, log.TxHash)
	price := initialze.SqrtPriceX96
	//tick := initialze.tick
	client, err := NewUniswapV3SimulatorCaller(log.Address, pm.rpc.GetClient())
	if err != nil {
		return nil, err
	}
	fee, err := client.Fee(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}
	tickSpacing, err := client.TickSpacing(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}
	token0, err := client.Token0(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}
	token1, err := client.Token1(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	pool := NewCorePoolFromConfig(PoolConfig{
		TickSpacing: tickSpacing.Int64(),
		Token0:      token0,
		Token1:      token1,
		Fee:         FeeAmount(fee.Int64()),
	})
	err = pool.Initialize(price)
	if err != nil {
		return nil, err
	}
	pm.pools[log.Address] = pool
	return pool, nil
}

// todo 应该先从磁盘加载snapshot, 然后以snapshot里的blockNum作为ingester的开始block
// snapshot定时持久化， 比如每10min持久化一次
func (pm *Simulator) HandleBlock(block *blockingester.NewBlockMsg) error {
	// 使用 block ingestor 的 transaction机制， 每个区块落盘一次
	// position和tick分表存放
	blockHash := block.Header.Hash()
	logs, err := pm.rpc.GetClient().FilterLogs(pm.ctx, ethereum.FilterQuery{
		BlockHash: &blockHash,
		Topics:    [][]common.Hash{{pm.InitializeID, pm.MintID, pm.BurnID, pm.SwapID}},
	})
	if err != nil {
		return err
	}
	//pm
	// 有变更的pool
	dirtyPool := map[string]*CorePool{}

	for _, log := range logs {
		topic0 := log.Topics[0]
		if topic0 == pm.InitializeID {
			pool, err := pm.InitPool(&log)
			if err != nil {
				logrus.Error(err)
			}
			dirtyPool[pool.PoolAddress] = pool
			return err
		} else if topic0 == pm.MintID {
			if pool, ok := pm.pools[log.Address]; !ok {
				logrus.Warnf("mint before initialize, tx: %s, pool: %s", log.TxHash, log.Address)
				continue
			} else {
				mint, err := parseUniv3MintEvent(&log)
				if err != nil {
					logrus.Warnf("failed parse mint event, tx: %s  pool: %s", log.TxHash, log.Address)
					continue
				}
				_, _, err = pool.Mint(mint.Owner, mint.TickLower, mint.TickUpper, mint.Amount)
				if err != nil {
					logrus.Errorf("failed execute mint event, %s tx: %s  pool: %s", err, log.TxHash, log.Address)
					return err
				}
			}
		} else if topic0 == pm.BurnID {
			if pool, ok := pm.pools[log.Address]; !ok {
				logrus.Warnf("burn before initialize, tx: %s, pool: %s", log.TxHash, log.Address)
				continue
			} else {
				burn, err := parseUniv3BurnEvent(&log)
				if err != nil {
					logrus.Warnf("failed parse burn event, tx: %s  pool: %s", log.TxHash, log.Address)
					continue
				}
				_, _, err = pool.Burn(burn.Sender, burn.TickLower, burn.TickUpper, burn.Amount)
				if err != nil {
					logrus.Errorf("failed execute burn event, %s tx: %s  pool: %s", err, log.TxHash, log.Address)
					return err
				}
			}
		} else if topic0 == pm.SwapID {
			if pool, ok := pm.pools[log.Address]; !ok {
				logrus.Warnf("swap before initialize, tx: %s, pool: %s", log.TxHash, log.Address)
				continue
			} else {
				swap, err := parseUniv3SwapEvent(&log)
				if err != nil {
					logrus.Warnf("failed parse swap event, tx: %s  pool: %s", log.TxHash, log.Address)
					continue
				}
				var amount decimal.Decimal
				if swap.Amount0.IsPositive() {
					amount
				}
				_, _, err = pool.handleSwap(swap.Amount0.IsPositive(), swap, swap.TickUpper, swap.Amount)
				if err != nil {
					logrus.Errorf("failed execute swap event, %s tx: %s  pool: %s", err, log.TxHash, log.Address)
					return err
				}
			}
		}
	}
	// pool变更落地
	err = pm.db.Transaction(func(tx *gorm.DB) error {
		for _, pool := range dirtyPool {
			err := pool.Flush(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// 从univ3创建区块开始同步所有pool
func (pm *Simulator) SyncToLatestAndListen() {
	pm.ingestor.Run()
}
