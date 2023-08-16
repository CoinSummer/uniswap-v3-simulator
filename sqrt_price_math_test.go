package uniswap_v3_simulator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var sqrtRatioX96 *big.Int

func init() {
	setString, _ := big.NewInt(0).SetString("79283619096159585451279", 10)
	sqrtRatioX96 = setString
}

func TestSqrtRatioX962number(t *testing.T) {
	type args struct {
		sqrtRatioX96 *big.Int
		price        *big.Int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test sqrtRatioX962number",
			args: args{
				sqrtRatioX96: sqrtRatioX96,
				price:        big.NewInt(1),
			},
			want: 1.0014,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			number := SqrtRatioX962Price(tt.args.sqrtRatioX96, tt.args.price)
			fmt.Printf("%g", float64(number.Int64())/1e6)
			assert.Equalf(t, tt.want, float64(SqrtRatioX962Price(tt.args.sqrtRatioX96, tt.args.price).Int64())/1e6, "SqrtRatioX962number(%v, %v)", tt.args.sqrtRatioX96, tt.args.price)
		})
	}
}
