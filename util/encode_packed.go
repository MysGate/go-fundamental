package util

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

func EncodePacked(input ...[]byte) []byte {
	return bytes.Join(input, nil)
}

func EncodeBytesString(v string) []byte {
	decoded, err := hex.DecodeString(v)
	if err != nil {
		panic(err)
	}
	return decoded
}

func EncodeUint256(v string) []byte {
	bn := new(big.Int)
	bn.SetString(v, 10)
	return math.U256Bytes(bn)
}

func EncodeBigInt(b *big.Int) []byte {
	return math.U256Bytes(b)
}

func EncodeUint256Array(arr []string) []byte {
	var res [][]byte
	for _, v := range arr {
		b := EncodeUint256(v)
		res = append(res, b)
	}

	return bytes.Join(res, nil)
}

func EncodeAddress(addr common.Address) []byte {
	return addr.Bytes()
}
