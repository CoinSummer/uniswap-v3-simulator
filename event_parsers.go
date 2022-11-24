package uniswap_v3_simulator

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
)

type UniV3InitializeEvent struct {
	RawEvent     *types.Log      `json:"raw_event"`
	SqrtPriceX96 decimal.Decimal `json:"sqrt_price_x96"`
	Removed      bool            `json:"removed"`
}
type UniV3SwapEvent struct {
	RawEvent     *types.Log      `json:"raw_event"`
	Sender       string          `json:"sender"`
	Amount0      decimal.Decimal `json:"amount0"`
	Amount1      decimal.Decimal `json:"amount1"`
	SqrtPriceX96 decimal.Decimal `json:"sqrt_price_x96"`
	Liquidity    decimal.Decimal `json:"Liquidity"`
	Recipient    string          `json:"to"`
	LogIndex     string          `json:"logIndex"`
	Removed      bool            `json:"removed"`
}

type UniV3MintEvent struct {
	RawEvent  *types.Log      `json:"raw_event"`
	Sender    string          `json:"sender"` // index value
	Owner     string          `json:"owner"`  // index value
	TickLower int             `json:"tick_lower"`
	TickUpper int             `json:"tick_upper"`
	Amount    decimal.Decimal `json:"amount"`
	Amount0   decimal.Decimal `json:"amount0"`
	Amount1   decimal.Decimal `json:"amount1"`
}
type UniV3BurnEvent struct {
	RawEvent  *types.Log      `json:"raw_event"`
	Owner     string          `json:"owner"` // index value
	TickLower int             `json:"tick_lower"`
	TickUpper int             `json:"tick_upper"`
	Amount    decimal.Decimal `json:"amount"`
	Amount0   decimal.Decimal `json:"amount0"`
	Amount1   decimal.Decimal `json:"amount1"`
}

var (
	int24, _   = abi.NewType("int24", "", nil)
	int256, _  = abi.NewType("int256", "", nil)
	uint160, _ = abi.NewType("uint160", "", nil)
	uint128, _ = abi.NewType("uint128", "", nil)
)

func parseUniv3SwapEvent(log *types.Log) (*UniV3SwapEvent, error) {
	event := log
	data := event.Data
	if len(event.Topics) != 3 {
		return nil, fmt.Errorf("topic not match,expect %d, got %d", 3, len(event.Topics))
	}
	amount0, ok := abi.ReadInteger(int256, data[0:32]).(*big.Int)
	if !ok {
		return nil, fmt.Errorf("parse swap err amount0 not a int")
	}

	amount1, ok := abi.ReadInteger(int256, data[32:32*2]).(*big.Int)
	if !ok {
		return nil, fmt.Errorf("parse swap err amount1 not a int")
	}
	price := abi.ReadInteger(uint160, data[32*2:32*3])
	liq := abi.ReadInteger(uint128, data[32*3:32*4])
	sqrtPriceX96, ok := price.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("parse swap err sqrtPriceX96 not a int")
	}
	liquidity, ok := liq.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("parse swap err Liquidity not a int")
	}

	parsed := &UniV3SwapEvent{
		RawEvent:     log,
		Amount0:      decimal.NewFromBigInt(amount0, 0),
		Amount1:      decimal.NewFromBigInt(amount1, 0),
		SqrtPriceX96: decimal.NewFromBigInt(sqrtPriceX96, 0),
		Liquidity:    decimal.NewFromBigInt(liquidity, 0),
	}
	// 看看xiaxin怎么做的。
	if parsed.Amount0.IsZero() && parsed.Amount1.IsZero() && parsed.Liquidity.IsZero() {
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
	tickLower, ok := abi.ReadInteger(int24, event.Topics[2].Bytes()).(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed read mint.tick_lower %s, tx: %s", tickLower, event.TxHash)
	}
	tickUpper, ok := abi.ReadInteger(int24, event.Topics[3].Bytes()).(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed read mint.tick_upper %s, tx: %s", tickUpper, event.TxHash)
	}
	parsed := &UniV3MintEvent{
		RawEvent:  log,
		Owner:     hash2Addr(event.Topics[1]),
		Sender:    common.BytesToAddress(data[:32]).Hex(),
		TickLower: int(tickLower.Int64()),
		TickUpper: int(tickUpper.Int64()),
		Amount:    decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32:32*2]), 0),
		Amount0:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*2:32*3]), 0),
		Amount1:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*3:32*4]), 0),
	}
	//if parsed.Amount0.IsZero() && parsed.Amount1.IsZero() {
	//	return nil, fmt.Errorf("mint amount0 and amount1 == 0")
	//}
	return parsed, nil
}
func parseUniv3BurnEvent(log *types.Log) (*UniV3BurnEvent, error) {
	event := log
	data := event.Data
	if len(event.Topics) != 4 {
		return nil, fmt.Errorf("topic not match,expect %d, got %d", 4, len(event.Topics))
	}
	tickLower, ok := abi.ReadInteger(int24, event.Topics[2].Bytes()).(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed read mint.tick_lower %s, tx: %s", tickLower, event.TxHash)
	}
	tickUpper, ok := abi.ReadInteger(int24, event.Topics[3].Bytes()).(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed read mint.tick_upper %s, tx: %s", tickUpper, event.TxHash)
	}
	parsed := &UniV3BurnEvent{
		RawEvent:  log,
		Owner:     hash2Addr(event.Topics[1]),
		TickLower: int(tickLower.Int64()),
		TickUpper: int(tickUpper.Int64()),
		Amount:    decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[:32]), 0),
		Amount0:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*1:32*2]), 0),
		Amount1:   decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[32*2:32*3]), 0),
	}
	//if parsed.Amount0.IsZero() && parsed.Amount1.IsZero() {
	//	return nil, fmt.Errorf("burn amount0 and amount1 == 0")
	//}
	return parsed, nil
}
func parseUniv3InitializeEvent(log *types.Log) (*UniV3InitializeEvent, error) {
	event := log
	data := event.Data
	if len(event.Topics) != 1 {
		return nil, fmt.Errorf("topic not match,expect %d, got %d", 4, len(event.Topics))
	}
	parsed := &UniV3InitializeEvent{
		RawEvent:     log,
		SqrtPriceX96: decimal.NewFromBigInt(big.NewInt(0).SetBytes(data[:32]), 0),
	}
	return parsed, nil
}
func hash2Addr(hs common.Hash) string {
	return strings.ToLower(common.BytesToAddress(hs[12:]).Hex())

}
