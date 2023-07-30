package util

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/MysGate/go-fundamental/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type EthAccount struct {
	privateKey string
	address    string
	publicKey  string
}

func (ea *EthAccount) ToString() string {
	return fmt.Sprintf("%s:%s", ea.address, ea.privateKey)
}

func NewEthAccount() *EthAccount {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		errMsg := fmt.Sprintf("NewEthAccount err:%+v", err)
		util.Logger().Error(errMsg)
		return nil
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Println(hexutil.Encode(privateKeyBytes)[2:])

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		util.Logger().Error("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return nil
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println(hexutil.Encode(publicKeyBytes)[4:])

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println(address)

	ea := &EthAccount{
		address:    address,
		publicKey:  hexutil.Encode(publicKeyBytes)[4:],
		privateKey: hexutil.Encode(privateKeyBytes)[2:],
	}
	return ea
}
