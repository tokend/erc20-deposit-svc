package eth

import "math/big"

var gweiPrecision = new(big.Int).SetInt64(1000000000)

func FromGwei(amount *big.Int) *big.Int {
	return new(big.Int).Mul(amount, gweiPrecision)
}

func ToGwei(amount *big.Int) *big.Int {
	return new(big.Int).Div(amount, gweiPrecision)
}
