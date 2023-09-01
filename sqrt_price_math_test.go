package uniswap_v3_simulator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var sqrtRatioX96 *big.Int
var sqrtRatioX961 *big.Int

func init() {
	setString, _ := big.NewInt(0).SetString("79283619096159585451279", 10)
	sqrtRatioX96 = setString

	setString1, _ := big.NewInt(0).SetString("170419876296389072947126", 10)
	sqrtRatioX961 = setString1
}

func TestSqrtRatioX962number(t *testing.T) {
	type args struct {
		sqrtRatioX96 *big.Int
		price        *big.Int
		decimals0    int
		decimals1    int
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
				decimals0:    18,
				decimals1:    6,
			},
			want: 1.0014,
		},
		{
			name: "uni",
			args: args{
				sqrtRatioX96: sqrtRatioX961,
				price:        big.NewInt(1),
				decimals0:    18,
				decimals1:    6,
			},
			want: 4.626806,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			number := SqrtRatioX962HumanPrice(tt.args.sqrtRatioX96, tt.args.price, tt.args.decimals0, tt.args.decimals1)
			fmt.Printf("%g", number)
			assert.Equalf(t, tt.want, SqrtRatioX962HumanPrice(tt.args.sqrtRatioX96, tt.args.price, tt.args.decimals0, tt.args.decimals1), "SqrtRatioX962number(%v, %v)", tt.args.sqrtRatioX96, tt.args.price)
		})
	}
}

func TestHumanPrice2SqrtRatioX96(t *testing.T) {
	type args struct {
		price     float64
		decimals0 int
		decimals1 int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				price:     1.0014,
				decimals0: 18,
				decimals1: 6,
			},
			want: "79283602830700258969327",
		},
		{
			name: "t2",
			args: args{
				price:     4.626806,
				decimals0: 18,
				decimals1: 6,
			},
			want: "170419869651098803081084",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x96, err := HumanPrice2SqrtRatioX96(tt.args.price, tt.args.decimals0, tt.args.decimals1)
			if err != nil {
				t.Errorf("exec HumanPrice2SqrtRatioX96 error: %s", err)
				return
			}

			assert.Equalf(t, tt.want, x96.String(), "HumanPrice2SqrtRatioX96(%v, %v, %v)", tt.args.price, tt.args.decimals0, tt.args.decimals1)
		})
	}
}
