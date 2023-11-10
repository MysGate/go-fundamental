package util

import (
	"fmt"
	"testing"
)

func Test_random(t *testing.T) {
	var one int
	var zero int
	for i := 0; i < 10000; i++ {
		a, err := GetRandomNumber(0, 1)
		if err != nil {
			fmt.Println(err)
			return
		}
		if a == 0 {
			zero++
		} else {
			one++
		}
		fmt.Println(a)
	}
	fmt.Println(zero)
	fmt.Println(one)
}
