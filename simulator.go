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
	"github.com/sirupsen/logrus"
	blockingester "gitlab.com/CoinSummer/Base/block-ingester"
	"gorm.io/gorm"
	"math/big"
	"strings"
)

var (
	ABI              = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount","type":"uint128"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"}],"name":"Burn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"address","name":"recipient","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount0","type":"uint128"},{"indexed":false,"internalType":"uint128","name":"amount1","type":"uint128"}],"name":"Collect","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint128","name":"amount0","type":"uint128"},{"indexed":false,"internalType":"uint128","name":"amount1","type":"uint128"}],"name":"CollectProtocol","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"paid0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"paid1","type":"uint256"}],"name":"Flash","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint16","name":"observationCardinalityNextOld","type":"uint16"},{"indexed":false,"internalType":"uint16","name":"observationCardinalityNextNew","type":"uint16"}],"name":"IncreaseObservationCardinalityNext","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"indexed":false,"internalType":"int24","name":"tick","type":"int24"}],"name":"Initialize","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount","type":"uint128"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"}],"name":"Mint","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint8","name":"feeProtocol0Old","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol1Old","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol0New","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol1New","type":"uint8"}],"name":"SetFeeProtocol","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"int256","name":"amount0","type":"int256"},{"indexed":false,"internalType":"int256","name":"amount1","type":"int256"},{"indexed":false,"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"indexed":false,"internalType":"uint128","name":"Liquidity","type":"uint128"},{"indexed":false,"internalType":"int24","name":"tick","type":"int24"}],"name":"Swap","type":"event"},{"inputs":[{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount","type":"uint128"}],"name":"burn","outputs":[{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount0Requested","type":"uint128"},{"internalType":"uint128","name":"amount1Requested","type":"uint128"}],"name":"collect","outputs":[{"internalType":"uint128","name":"amount0","type":"uint128"},{"internalType":"uint128","name":"amount1","type":"uint128"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint128","name":"amount0Requested","type":"uint128"},{"internalType":"uint128","name":"amount1Requested","type":"uint128"}],"name":"collectProtocol","outputs":[{"internalType":"uint128","name":"amount0","type":"uint128"},{"internalType":"uint128","name":"amount1","type":"uint128"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"factory","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"fee","outputs":[{"internalType":"uint24","name":"","type":"uint24"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"feeGrowthGlobal0X128","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"feeGrowthGlobal1X128","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"flash","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint16","name":"observationCardinalityNext","type":"uint16"}],"name":"increaseObservationCardinalityNext","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"Liquidity","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"maxLiquidityPerTick","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount","type":"uint128"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"mint","outputs":[{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"observations","outputs":[{"internalType":"uint32","name":"blockTimestamp","type":"uint32"},{"internalType":"int56","name":"tickCumulative","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityCumulativeX128","type":"uint160"},{"internalType":"bool","name":"initialized","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint32[]","name":"secondsAgos","type":"uint32[]"}],"name":"observe","outputs":[{"internalType":"int56[]","name":"tickCumulatives","type":"int56[]"},{"internalType":"uint160[]","name":"secondsPerLiquidityCumulativeX128s","type":"uint160[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"Positions","outputs":[{"internalType":"uint128","name":"Liquidity","type":"uint128"},{"internalType":"uint256","name":"FeeGrowthInside0LastX128","type":"uint256"},{"internalType":"uint256","name":"FeeGrowthInside1LastX128","type":"uint256"},{"internalType":"uint128","name":"TokensOwed0","type":"uint128"},{"internalType":"uint128","name":"TokensOwed1","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"protocolFees","outputs":[{"internalType":"uint128","name":"token0","type":"uint128"},{"internalType":"uint128","name":"token1","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint8","name":"feeProtocol0","type":"uint8"},{"internalType":"uint8","name":"feeProtocol1","type":"uint8"}],"name":"setFeeProtocol","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"slot0","outputs":[{"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"internalType":"int24","name":"tick","type":"int24"},{"internalType":"uint16","name":"observationIndex","type":"uint16"},{"internalType":"uint16","name":"observationCardinality","type":"uint16"},{"internalType":"uint16","name":"observationCardinalityNext","type":"uint16"},{"internalType":"uint8","name":"feeProtocol","type":"uint8"},{"internalType":"bool","name":"unlocked","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"}],"name":"snapshotCumulativesInside","outputs":[{"internalType":"int56","name":"tickCumulativeInside","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityInsideX128","type":"uint160"},{"internalType":"uint32","name":"secondsInside","type":"uint32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"bool","name":"zeroForOne","type":"bool"},{"internalType":"int256","name":"amountSpecified","type":"int256"},{"internalType":"uint160","name":"sqrtPriceLimitX96","type":"uint160"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"swap","outputs":[{"internalType":"int256","name":"amount0","type":"int256"},{"internalType":"int256","name":"amount1","type":"int256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"int16","name":"","type":"int16"}],"name":"tickBitmap","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"tickSpacing","outputs":[{"internalType":"int24","name":"","type":"int24"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"int24","name":"","type":"int24"}],"name":"Ticks","outputs":[{"internalType":"uint128","name":"liquidityGross","type":"uint128"},{"internalType":"int128","name":"liquidityNet","type":"int128"},{"internalType":"uint256","name":"feeGrowthOutside0X128","type":"uint256"},{"internalType":"uint256","name":"feeGrowthOutside1X128","type":"uint256"},{"internalType":"int56","name":"tickCumulativeOutside","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityOutsideX128","type":"uint160"},{"internalType":"uint32","name":"secondsOutside","type":"uint32"},{"internalType":"bool","name":"initialized","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`
	TOPIC_INITIALIZE = common.HexToHash("0x98636036cb66a9c19a37435efc1e90142190214e8abeb821bdba3f2990dd4c95")
	TOPIC_BURN       = common.HexToHash("0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c")
	TOPIC_SWAP       = common.HexToHash("0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67")
	TOPIC_MINT       = common.HexToHash("0x7a53080ba414158be7ec69b987b5fb7d07dee101fe85488f0853ae16239d0bde")
)

var (
	skipAddress = []common.Address{common.HexToAddress("0xa87998484c19d68807debdc280e18424d55743a9"), common.HexToAddress("0xcba27c8e7115b4eb50aa14999bc0866674a96ecb"), common.HexToAddress("0x979f63b8279376ef8205fb536b16080cd1d45058")}
)

type Simulator struct {
	pools        map[common.Address]*CorePool
	dirtyPools   map[string]*CorePool
	Abi          abi.ABI
	InitializeID common.Hash
	MintID       common.Hash
	BurnID       common.Hash
	SwapID       common.Hash
	ingestor     *blockingester.BlockIngester
	rpc          *blockingester.EthRpcClientPool
	wss          *ethclient.Client
	db           *gorm.DB
	dbfile       string
	ctx          context.Context
}

// SYNC from univ3 created
func NewPoolManager(dbFile string, wss string, rpcs []string) *Simulator {
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
		pools:      map[common.Address]*CorePool{},
		dirtyPools: map[string]*CorePool{},
		rpc:        rpc,
		wss:        wssClient,
		db:         db,
		dbfile:     dbFile,
		ctx:        context.Background(),
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

	db.AutoMigrate(&CorePool{})

	var currentPool []*CorePool
	err = db.Find(&currentPool).Error
	if err != nil {
		logrus.Fatal(err)
	}
	for _, pool := range currentPool {
		pm.pools[common.HexToAddress(pool.PoolAddress)] = pool
	}
	return pm
}

func (pm *Simulator) CurrentBlock() uint64 {
	return pm.ingestor.CurrentBlock()
}

func (pm *Simulator) InitPool(log *types.Log) (*CorePool, error) {
	if _, exist := pm.pools[log.Address]; exist {
		return nil, fmt.Errorf("pool exists %s", log.Address)
	}

	initialze, err := parseUniv3InitializeEvent(log)
	if err != nil {
		return nil, err
	}

	logrus.Infof("initialize pool: %s,  tx: %s, price: %s", log.Address, log.TxHash, initialze.SqrtPriceX96)
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

	pool := NewCorePoolFromConfig(log.Address.String(), PoolConfig{
		TickSpacing: tickSpacing.Int64(),
		Token0:      token0,
		Token1:      token1,
		Fee:         FeeAmount(fee.Int64()),
	})
	err = pool.Initialize(price)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (pm *Simulator) HandleLogs(logs []types.Log) error {
	// 有变更的pool
	for _, log := range logs {
		if log.Address == skipAddress[0] || log.Address == skipAddress[1] || log.Address == skipAddress[2] {
			continue
		}
		topic0 := log.Topics[0]
		if topic0 == pm.InitializeID {
			pool, err := pm.InitPool(&log)
			if err != nil {
				logrus.Error(err)
			}
			pool.DeployBlockNum = log.BlockNumber
			pool.CurrentBlockNum = log.BlockNumber
			pm.dirtyPools[pool.PoolAddress] = pool
			pm.pools[log.Address] = pool
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
				//s, _ := json.Marshal(mint)
				//logrus.Infof("mint: %s %s %s", log.Address, log.TxHash, string(s))
				_, _, err = pool.Mint(mint.Owner, mint.TickLower, mint.TickUpper, mint.Amount)
				if err != nil {
					logrus.Errorf("failed execute mint event, %s tx: %s  pool: %s", err, log.TxHash, log.Address)
					return err
				}
				pool.CurrentBlockNum = log.BlockNumber
				pm.dirtyPools[pool.PoolAddress] = pool
			}
		} else if topic0 == pm.BurnID {
			if pool, ok := pm.pools[log.Address]; !ok {
				logrus.Warnf("burn before initialize, tx: %s, pool: %s", log.TxHash, log.Address)
				continue
			} else {
				burn, err := parseUniv3BurnEvent(&log)
				if err != nil {
					logrus.Warnf("failed parse burn event, tx: %s  pool: %s err: %s", log.TxHash, log.Address, err)
					continue
				}
				//s, _ := json.Marshal(burn)
				//logrus.Infof("burn: %s %s %s", log.Address, log.TxHash, string(s))
				_, _, err = pool.Burn(burn.Owner, burn.TickLower, burn.TickUpper, burn.Amount)
				if err != nil {
					logrus.Errorf("failed execute burn event, %s tx: %s  pool: %s", err, log.TxHash, log.Address)
					return err
				}
				pool.CurrentBlockNum = log.BlockNumber
				pm.dirtyPools[pool.PoolAddress] = pool
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
				//s, _ := json.Marshal(swap)
				//logrus.Infof("swap: %s %s %s", log.Address, log.TxHash, string(s))
				amountSpecified, sqrtPriceX96, err := pool.ResolveInputFromSwapResultEvent(swap)
				if err != nil {
					logrus.Fatalf("failed resolve swap param from event, tx: %s  pool: %s, %s", log.TxHash, log.Address, err)
				}

				_, _, _, err = pool.HandleSwap(swap.Amount0.IsPositive(), amountSpecified, sqrtPriceX96, false)
				if err != nil {
					logrus.Fatalf("failed execute swap event, %s tx: %s  pool: %s", err, log.TxHash, log.Address)
				}
				pool.CurrentBlockNum = log.BlockNumber
				pm.dirtyPools[pool.PoolAddress] = pool
			}
		}
	}
	return nil
}

func (pm *Simulator) MaxSyncedBlockNum() (uint64, error) {
	var lastBlock *uint64
	err := pm.db.Model(&CorePool{}).Select("max(current_block_num) as last_block").Scan(&lastBlock).Error
	if err != nil {
		return 0, err
	}
	if lastBlock == nil {
		return 0, nil
	}
	return *lastBlock, nil
}

func (pm *Simulator) FlushPools() error {
	// pool变更落地
	err := pm.db.Transaction(func(tx *gorm.DB) error {
		for _, pool := range pm.dirtyPools {
			err := pool.Flush(tx)
			if err != nil {
				logrus.Errorf("failed flush pool %s", err)
				return err
			}
			logrus.Infof("flush pool: %s", pool.PoolAddress)
		}
		return nil
	})
	if err != nil {
		logrus.Warnf("failed save snapshot %s", err)
		return err
	} else {
		pm.dirtyPools = map[string]*CorePool{}
		return nil
	}
}

// end is inclusive
func (pm *Simulator) SyncHistory(step uint64) (uint64, error) {
	// 从数据库获取start, max(currentBlock)
	lastBlock, err := pm.MaxSyncedBlockNum()
	if err != nil {
		return 0, err
	}
	if lastBlock == 0 {
		// univ3 factory deploy
		lastBlock = 12369620
	}
	start := lastBlock + 1
	latest, err := pm.rpc.GetClient().BlockNumber(pm.ctx)
	if err != nil {
		return 0, err
	}
	end := latest
	flushStep := 0

	for {
		if start > end {
			return end, nil
		}
		flushStep += 1
		var minEnd uint64
		if start+step > end {
			minEnd = end
		} else {
			minEnd = start + step
		}
		logrus.Infof("sync blocks: %d - %d", start, minEnd)
		logs, err := pm.rpc.GetClient().FilterLogs(pm.ctx, ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(start)),
			ToBlock:   big.NewInt(int64(minEnd)),
			Topics:    [][]common.Hash{{pm.InitializeID, pm.MintID, pm.BurnID, pm.SwapID}},
			//Addresses: []common.Address{common.HexToAddress("0xCba27C8e7115b4Eb50Aa14999BC0866674a96eCB")},
		})
		if err != nil {
			return 0, err
		}
		err = pm.HandleLogs(logs)
		if err != nil {
			return 0, err
		}
		// 每10w block flush一次
		if flushStep%10 == 0 {
			err = pm.FlushPools()
			if err != nil {
				return 0, err
			}
		}
		start = minEnd + 1

	}

}

// todo 应该先从磁盘加载snapshot, 然后以snapshot里的blockNum作为ingester的开始block
// snapshot定时持久化， 比如每10min持久化一次
func (pm *Simulator) HandleBlock(block *blockingester.NewBlockMsg) error {
	logrus.Infof("sync to block: %s", block.Header.Number)
	//return nil
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
	return pm.HandleLogs(logs)
}

// 从univ3创建区块开始同步所有pool
func (pm *Simulator) Run() error {
	syncTo, err := pm.SyncHistory(10000)
	if err != nil {
		return err
	}
	err = pm.FlushPools()
	if err != nil {
		return err
	}
	ingesterDB, err := gorm.Open(sqlite.Open(pm.dbfile), &gorm.Config{})
	if err != nil {
		logrus.Fatal(err)
	}
	// 旧数据自行同步了， 无需ingester
	err = pm.db.Exec("drop table table_block_ingesters").Error
	if err != nil {
		return err
	}

	ingester := blockingester.LoadOrCreateBlockIngester("arbitary", ingesterDB, pm.wss, big.NewInt(int64(syncTo+1)), true, true, pm, context.Background())
	pm.ingestor = ingester
	pm.ingestor.Run()
	return nil
}

func (pm *Simulator) ForkPool(blockNum uint64, poolAddress string) (*CorePool, error) {

	if pool, ok := pm.pools[common.HexToAddress(poolAddress)]; !ok {
		return nil, fmt.Errorf("pool not exists %s", poolAddress)
	} else {
		if pm.CurrentBlock() != blockNum {
			return nil, fmt.Errorf("simulation req at %d , but simulator's block at %d", blockNum, pm.CurrentBlock())
		}
		fork := pool.Clone()
		return fork, nil
	}
}
