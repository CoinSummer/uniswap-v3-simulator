package uniswap_v3_simulator

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	blockingester "gitlab.com/CoinSummer/Base/block-ingester"
	"gorm.io/gorm"
	"math/big"
)

type Simulator struct {
	pools    map[string]*CorePool
	ingestor *blockingester.BlockIngester
	rpc      *blockingester.EthRpcClientPool
	wss      *ethclient.Client
	ctx      context.Context
}

func NewPoolManager(startBlock int64, db *gorm.DB, wss string, rpc *blockingester.EthRpcClientPool) *Simulator {
	wssClient, err := ethclient.Dial(wss)
	if err != nil {
		panic(err)
	}
	pm := &Simulator{
		pools: map[string]*CorePool{},
		rpc:   rpc,
		wss:   wssClient,
		ctx:   context.Background(),
	}
	ingester := blockingester.LoadOrCreateBlockIngester("arbitary", db, wssClient, big.NewInt(startBlock), true, true, pm, context.Background())
	pm.ingestor = ingester
	return pm
}

func (pm *Simulator) HandleBlock(block *blockingester.NewBlockMsg) error {
	// 使用 block ingestor 的 transaction机制， 每个区块落盘一次
	blockHash := block.Header.Hash()
	logs, err := pm.rpc.GetClient().FilterLogs(pm.ctx, ethereum.FilterQuery{
		BlockHash: &blockHash,
		Topics:    [][]common.Hash{{common.HexToHash("0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67")}},
	})
	if err != nil {
		return err
	}
	for _, i := range logs {
		fmt.Println(i)
	}
	return nil
}

// 从univ3创建区块开始同步所有pool
func (pm *Simulator) Start() {
	pm.ingestor.Run()
}
