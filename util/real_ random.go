package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

// gen  min<= a <= max
func GetRandomNumber(min, max int) (int, error) {
	if min < 0 || max < 0 {
		err := errors.New("min, max must over 0")
		return 0, err
	}

	if min == max {
		return min, nil
	}

	if max < min {
		err := errors.New("min must low max")
		return 0, err
	}

	bigMax := big.NewInt(int64(max + 1))
	for {
		result, err := rand.Int(rand.Reader, bigMax)
		if err != nil {
			continue
		}
		number := result.String()
		num, err := strconv.Atoi(number)
		if err != nil {
			fmt.Println(num)
			continue
		}
		if num >= min && num <= max {
			return num, nil
		}
	}
}
