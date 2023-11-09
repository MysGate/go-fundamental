package signature

import (
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// HashTypedData is used to calculate the hash of EIP-712 conformant typed data
// hash = keccak256("\x19${byteVersion}${domainSeparator}${hashStruct(message)}")
func HashTypedData(data apitypes.TypedData) ([]byte, []byte, error) {
	domainSeparator, err := data.HashStruct("EIP712Domain", data.Domain.Map())
	if err != nil {
		return nil, nil, err
	}
	dataHash, err := data.HashStruct(data.PrimaryType, data.Message)
	if err != nil {
		return nil, nil, err
	}
	prefixedData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(dataHash)))
	prefixedDataHash := crypto.Keccak256(prefixedData)
	return dataHash, prefixedDataHash, nil
}

// RecoveryTypedDataAddressEx is used to recover the signer address of the TypedData signature
func RecoveryTypedDataAddressEx(data apitypes.TypedData, signature []byte) (ethcommon.Address, error) {
	_, dataHash, err := HashTypedData(data)
	if err != nil {
		return ethcommon.Address{}, err
	}
	return RecoveryAddressEx(dataHash, signature)
}

// VerifyTypedDataSignatureEx is used to verify the signer address of the TypedData signature
func VerifyTypedDataSignatureEx(address ethcommon.Address, data apitypes.TypedData, signature []byte) (bool, error) {
	recoveredAddress, err := RecoveryTypedDataAddressEx(data, signature)
	if err != nil {
		return false, err
	}
	return recoveredAddress == address, nil
}

// VerifyTypedDataHexSignatureEx is used to verify the signer address of the TypedData signature
func VerifyTypedDataHexSignatureEx(address ethcommon.Address, data apitypes.TypedData, signature string) (bool, error) {
	sig, err := HexDecode(signature)
	if err != nil {
		return false, err
	}
	recoveredAddress, err := RecoveryTypedDataAddressEx(data, sig)
	if err != nil {
		return false, err
	}
	return recoveredAddress == address, nil
}
