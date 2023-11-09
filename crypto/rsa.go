package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
)

func GenRsaKey(bits int) (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", "", err
	}

	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	privateStr := hex.EncodeToString(derStream)

	pubStr := hex.EncodeToString(derPkix)
	return privateStr, pubStr, nil
}

func RsaEncrypt(data []byte, pubKey string) ([]byte, error) {
	pubBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}

	pubInterface, err := x509.ParsePKIXPublicKey(pubBytes)
	if err != nil {
		return nil, err
	}

	pub := pubInterface.(*rsa.PublicKey)
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func RsaDecrypt(ciphertext []byte, privateKey string) ([]byte, error) {
	privateBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}

	priv, err := x509.ParsePKCS1PrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}

	data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		return nil, err
	}
	return data, nil
}
