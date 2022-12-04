package uniswap_v3_simulator

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// 分叉而不影响原数据
type SimulatorFork struct {
	Pools     map[common.Address]*CorePool
	simulator *Simulator
}

func NewSimulatorSnapshot(s *Simulator) *SimulatorFork {
	return &SimulatorFork{
		Pools:     map[common.Address]*CorePool{},
		simulator: s,
	}
}

func (s *SimulatorFork) GetPool(addr common.Address) (*CorePool, error) {
	if _, ok := s.Pools[addr]; !ok {
		// fork
		forkedPool, err := s.simulator.ForkPool(addr)
		if err != nil {
			return nil, err
		}
		s.Pools[addr] = forkedPool
	}
	return s.Pools[addr], nil
}

func (s *SimulatorFork) HandleLogs(logs []types.Log) error {
	for _, log := range logs {
		if log.Address == skipAddress[0] || log.Address == skipAddress[1] || log.Address == skipAddress[2] {
			continue
		}
		if len(log.Topics) == 0 {
			return nil
		}
		topic0 := log.Topics[0]
		if topic0 == s.simulator.InitializeID {
			pool, err := s.simulator.NewPool(&log)
			if err != nil {
				logrus.Error(err)
				if err.Error() == "execution reverted" {
					continue
				} else {
					logrus.Fatal(err)
				}
			}
			pool.DeployBlockNum = log.BlockNumber
			pool.CurrentBlockNum = log.BlockNumber
			s.Pools[common.HexToAddress(pool.PoolAddress)] = pool
		} else {
			var pool *CorePool
			var err error
			if topic0 == s.simulator.MintID || topic0 == s.simulator.BurnID || topic0 == s.simulator.SwapID {
				pool, err = s.GetPool(log.Address)
				if err != nil {
					return err
				}
			} else {
				continue
			}

			if topic0 == s.simulator.MintID {
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
				pool.CurrentBlockNum = log.BlockNumber
			} else if topic0 == s.simulator.BurnID {
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
			} else if topic0 == s.simulator.SwapID {
				swap, err := parseUniv3SwapEvent(&log)
				if err != nil {
					logrus.Warnf("failed parse swap event, tx: %s  pool: %s", log.TxHash, log.Address)
					continue
				}
				amountSpecified, sqrtPriceX96, err := pool.ResolveInputFromSwapResultEvent(swap)
				if err != nil {
					s, _ := json.Marshal(swap)
					logrus.Infof("swap: %s %s %s", log.Address, log.TxHash, string(s))
					return err
				}

				_, _, _, err = pool.HandleSwap(swap.Amount0.IsPositive(), amountSpecified, sqrtPriceX96, false)
				if err != nil {
					logrus.Errorf("failed execute swap event, %s tx: %s  pool: %s", err, log.TxHash, log.Address)
					return err
				}
				pool.CurrentBlockNum = log.BlockNumber
			}
		}
	}
	return nil
}
