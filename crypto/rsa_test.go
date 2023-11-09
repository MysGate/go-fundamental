package crypto

import (
	"fmt"
	"testing"
)

func Test_rsa_gen(t *testing.T) {
	priStr, pubStr, err := GenRsaKey(1024)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("private string:")
	fmt.Println(priStr)
	fmt.Println("public string:")
	fmt.Println(pubStr)
}

func Test_rsa_encrypto_decrypto(t *testing.T) {
	fmt.Println("gen keys")
	priStr, pubStr, err := GenRsaKey(1024)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("encrypto text")
	plainText := "Hello World"
	encrypto, err := RsaEncrypt([]byte(plainText), pubStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(encrypto))
	fmt.Println("decrypto text")
	plainBytes, err := RsaDecrypt(encrypto, priStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(plainBytes))
}
