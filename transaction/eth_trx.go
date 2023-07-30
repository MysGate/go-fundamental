package transaction

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/MysGate/go-fundamental/util"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CreateTransactionOpts(client *ethclient.Client, key *ecdsa.PrivateKey, chainId uint64, caller common.Address) (opts *bind.TransactOpts, err error) {
	nonce, err := client.PendingNonceAt(context.Background(), caller)
	if err != nil {
		errMsg := fmt.Sprintf("CreateTransactionOpts:client.PendingNonceAt err: %+v", err)
		util.Logger().Error(errMsg)
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("CreateTransactionOpts:client.SuggestGasPrice err: %+v", err)
		util.Logger().Error(errMsg)
		return nil, err
	}

	srcChainID := big.NewInt(int64(chainId))
	opts, err = bind.NewKeyedTransactorWithChainID(key, srcChainID)
	if err != nil {
		errMsg := fmt.Sprintf("CreateTransactionOpts:NewKeyedTransactorWithChainID err: %+v", err)
		util.Logger().Error(errMsg)
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
		util.Logger().Error(errMsg)
		return nil, false, err
	}

	return receipt, receipt.Status == types.ReceiptStatusSuccessful, nil
}
