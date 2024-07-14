package main

import (
	"errors"
	"fmt"
	"math/rand"
)

type randNumber struct{}

func (r randNumber) calculate(options Options, cacheI cacheItem) (any, *cacheItem, error) {
	miN := options.InitialBounds.Min
	maX := options.InitialBounds.Max

	if maX < miN {
		return "", &cacheItem{}, errors.New(
			fmt.Sprintf("maximum value %d can not be less then minimum value %d", maX, miN))
	}

	if miN == maX {
		return miN, nil, nil
	}

	return int64(rand.Intn(maX-miN) + miN), nil, nil
}
