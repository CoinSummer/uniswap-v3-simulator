package uniswap_v3_simulator

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"math/big"
)

type UniV3SwapEvent struct {
	RawEvent  *types.Log      `json:"raw_event"`
	Sender    string          `json:"sender"`
	Amount0   decimal.Decimal `json:"amount0In"`
	Amount1   decimal.Decimal `json:"amount1In"`
	Recipient string          `json:"to"`
	LogIndex  string          `json:"logIndex"`
	Removed   bool            `json:"removed"`
}

type UniV3MintEvent struct {
	RawEvent  *types.Log      `json:"raw_event"`
	Sender    string          `json:"sender"` // index value
	Owner     string          `json:"owner"`  // index value
	TickLower decimal.Decimal `json:"tick_lower"`
	TickUpper decimal.Decimal `json:"tick_upper"`
	Amount    decimal.Decimal `json:"amount"`
	Amount0   decimal.Decimal `json:"amount0"`
	Amount1   decimal.Decimal `json:"amount1"`
}
type UniV3BurnEvent struct {
	RawEvent  *types.Log      `json:"raw_event"`
	Sender    string          `json:"sender"` // index value
	TickLower decimal.Decimal `json:"tick_lower"`
	TickUpper decimal.Decimal `json:"tick_upper"`
	Amount    decimal.Decimal `json:"amount"`
	Amount0   decimal.Decimal `json:"amount0"`
	Amount1   decimal.Decimal `json:"amount1"`
}

func parseUniv3SwapEvent(log *types.Log) (*UniV3SwapEvent, error) {
	event := log
	data := event.Data
	if len(event.Topics) != 3 {
		return nil, fmt.Errorf("topic not match,expect %d, got %d", 3, len(event.Topics))
	}
	int256, _ := abi.NewType("int256", "", nil)
	a0 := abi.ReadInteger(int256, data[0:32])

	amount0, ok := a0.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("parse swap err amount0 not a int")
	}

	a1 := abi.ReadInteger(int256, data[32:32*2])

	amount1, ok := a1.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("parse swap err amount1 not a int")
	}

	parsed := &UniV3SwapEvent{
		RawEvent: log,
		Amount0:  decimal.NewFromBigInt(amount0, 0),
		Amount1:  decimal.NewFromBigInt(amount1, 0),
	}
	if parsed.Amount0.IsZero() && parsed.Amount1.IsZero() {
		return nil, fmt.Errorf("swap amoun is 0: %s", log.TxHash)
	}
	return parsed, nil
}
func parseUniv3MintEvent(log *types.Log) (*UniV3MintEvent, error) {
	event := log
	data := event.Data
	if len(event.Topics) != 4 {
		return nil, fmt.Errorf("topic not match,expect %d, got %d", 4, len(event.Topics))
	}
	parsed := &UniV3MintEvent{
		RawEvent: log,
		//Owner:     hash2Addr(event.Topics[1]),
		Sender:    common.BytesToAddress(data[:32]).Hex(),
		TickLower: decimal.NewFromBigInt(big.NewInt(0).SetBytes(event.Topics[2].Bytes()), 0),
		TickUpper: decimal.NewFromBigInt(big.NewInt(0).SetBytes(event.Topics[3].Bytes()), 0),
		Amount:    decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32:32*2]), 0),
		Amount0:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*2:32*3]), 0),
		Amount1:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*3:32*4]), 0),
	}
	if parsed.Amount0.IsZero() && parsed.Amount1.IsZero() {
		return nil, fmt.Errorf("mint amount0 and amount1 == 0")
	}
	return parsed, nil
}
func parseUniv3BurnEvent(log *types.Log) (*UniV3BurnEvent, error) {
	event := log
	data := event.Data
	if len(event.Topics) != 4 {
		return nil, fmt.Errorf("topic not match,expect %d, got %d", 4, len(event.Topics))
	}
	parsed := &UniV3BurnEvent{
		RawEvent:  log,
		TickLower: decimal.NewFromBigInt(big.NewInt(0).SetBytes(event.Topics[2].Bytes()), 0),
		TickUpper: decimal.NewFromBigInt(big.NewInt(0).SetBytes(event.Topics[3].Bytes()), 0),
		Amount:    decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[:32]), 0),
		Amount0:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*1:32*2]), 0),
		Amount1:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*2:32*3]), 0),
	}
	if parsed.Amount0.IsZero() && parsed.Amount1.IsZero() {
		return nil, fmt.Errorf("burn amount0 and amount1 == 0")
	}
	return parsed, nil
}
