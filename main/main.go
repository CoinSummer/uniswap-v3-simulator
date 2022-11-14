package main

import uniswap_v3_simulator "uniswap-v3-simulator"

func main() {
	smt := uniswap_v3_simulator.NewPoolManager(12369620, "simulator.db", "wss://mainnet.infura.io/ws/v3/81e90c9cd6a0430182e3a2bec37f2ba0", []string{"https://eth-hk1.csnodes.com/v1/973eeba6738a7d8c3bd54f91adcbea89"})

	smt.SyncToLatestAndListen()

}
