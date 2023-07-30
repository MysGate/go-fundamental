package account

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/MysGate/go-fundamental/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type EthAccount struct {
	PrivateKey string
	Address    string
	PublicKey  string
}

func (ea *EthAccount) ToString() string {
	return fmt.Sprintf("%s:%s", ea.Address, ea.PrivateKey)
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
		Address:    address,
		PublicKey:  hexutil.Encode(publicKeyBytes)[4:],
		PrivateKey: hexutil.Encode(privateKeyBytes)[2:],
	}
	return ea
}
