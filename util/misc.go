package util

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"

	"github.com/MysGate/go-fundamental/core"
	"github.com/bwmarrin/snowflake"
	"github.com/shopspring/decimal"
)

var IsAlphanumeric = regexp.MustCompile(`^[0-9a-zA-Z]+$`).MatchString

func ConvertHexToDecimalInStringFormat(hexString string) string {
	i := new(big.Int)
	// if hexString with '0x' prefix, using fmt.Sscan()
	fmt.Sscan(hexString, i)
	// if hexString without '0x' prefix, using i.SetString()
	//i.SetString(hexString, 16)

	return fmt.Sprintf("%v", i)
}

func ConvertFloat64ToTokenAmount(amount float64, decimals int) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(amount)

	fp := math.Pow10(decimals)

	coin := new(big.Float)
	coin.SetInt(big.NewInt(int64(fp)))
	bigval.Mul(bigval, coin)

	result := new(big.Int)
	bigval.Int(result) // store converted number in result

	return result
}

func PadLeft(str, pad string, length int) string {
	for {
		str = pad + str
		if len(str) >= length {
			return str[0:length]
		}
	}
}

func IsAnAddress(address string) bool {
	return len(address) == core.AddressFixedLength+2 && address[:2] == "0x" && IsAlphanumeric(address)
}

func IsValidTxHash(txHash string) bool {
	return len(txHash) == core.TxHashFixedLength && txHash[:2] == "0x" && IsAlphanumeric(txHash)
}

func ConvertTokenAmountToFloat64(amt string, tokenDecimal int32) float64 {
	amount, _ := decimal.NewFromString(amt)
	amount_converted := amount.Div(decimal.New(1, tokenDecimal))
	amountFloat, _ := amount_converted.Float64()
	return amountFloat
}

func ConvertBigIntFromString(v0, v1 string) (n0 *big.Int, n1 *big.Int, err error) {
	n0 = new(big.Int)
	n0, ok := n0.SetString(v0, 10)
	if !ok {
		err = errors.New("RawProofToZkProof err")
		Logger().Error(err.Error())
		return
	}

	n1 = new(big.Int)
	n1, ok = n1.SetString(v1, 10)
	if !ok {
		err = errors.New("RawProofToZkProof err")
		Logger().Error(err.Error())
		return
	}
	return
}

func GenerateIncreaseID() (int64, error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		Logger().Error(fmt.Sprintf("GenerateIncreaseID err:%+v", err))
		return 0, err
	}
	// Generate a snowflake ID.
	id := node.Generate()

	return id.Int64(), nil
}
