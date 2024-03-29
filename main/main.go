package main

import uniswap_v3_simulator "github.com/CoinSummer/uniswap-v3-simulator"

func main() {
	smt := uniswap_v3_simulator.NewPoolManager("simulator.db", "https://eth-hk1.csnodes.com/v1/973eeba6738a7d8c3bd54f91adcbea89", 12369620)

	//err := smt.Init(10000)
	_, err := smt.SyncTo(16381994, 10000)
	if err != nil {
		panic(err)
	}
	smt.FlushPools()

}
