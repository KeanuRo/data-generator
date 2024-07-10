package main

import (
	"math/rand"
)

type randNumber struct{}

func (r randNumber) calculate(options Options, linkedId, generatorId, iteration int, variable string) (any, error) {
	miN := 10
	maX := 30
	return rand.Intn(maX-miN) + miN, nil
}
