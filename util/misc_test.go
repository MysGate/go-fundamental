package util

import (
	"fmt"
	"testing"
)

func TestRemoveSlice(t *testing.T) {
	var a []int
	a = append(a, 0)
	a = append(a, 11)
	a = append(a, 31)
	a = append(a, 21)

	b := RemoveIndex(a, 3)
	fmt.Print(b)
}

func TestPrivateToAddress(t *testing.T) {
	key := "1c4b660c56987ea731b4c894f4bf8d374b3a4e60b963dc5e61d326517cd0cc2f"
	addr, _ := PrivateToAddress(key)
	fmt.Println(addr)
}
