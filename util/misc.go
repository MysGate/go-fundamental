package util

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"

	"github.com/MysGate/go-fundamental/core"
	"github.com/bwmarrin/snowflake"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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

func RemoveIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func CreateTransactionOpts(client *ethclient.Client, key *ecdsa.PrivateKey, chainId uint64, caller common.Address) (opts *bind.TransactOpts, err error) {
	nonce, err := client.PendingNonceAt(context.Background(), caller)
	if err != nil {
		errMsg := fmt.Sprintf("CreateTransactionOpts:client.PendingNonceAt err: %+v", err)
		Logger().Error(errMsg)
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("CreateTransactionOpts:client.SuggestGasPrice err: %+v", err)
		Logger().Error(errMsg)
		return nil, err
	}

	srcChainID := big.NewInt(int64(chainId))
	opts, err = bind.NewKeyedTransactorWithChainID(key, srcChainID)
	if err != nil {
		errMsg := fmt.Sprintf("CreateTransactionOpts:NewKeyedTransactorWithChainID err: %+v", err)
		Logger().Error(errMsg)
		return nil, err
	}

	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = big.NewInt(0) // in wei
	opts.GasLimit = uint64(0)  // in units
	opts.GasPrice = new(big.Int).Mul(gasPrice, big.NewInt(2))

	return opts, nil
}

func TxWaitToSync(ctx context.Context, client *ethclient.Client, tx *types.Transaction) (*types.Receipt, bool, error) {
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		errMsg := fmt.Sprintf("TxWaitToSync:bind.WaitMine err: %+v", err)
		Logger().Error(errMsg)
		return nil, false, err
	}

	return receipt, receipt.Status == types.ReceiptStatusSuccessful, nil
}

func PrivateToAddress(key string) (string, error) {
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", err
	}
	addr := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return addr, nil
}
