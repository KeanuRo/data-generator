package main

type rand struct{}

func (r rand) calculate(options Options, linkedId, generatorId, iteration int, variable string) (any, error) {
	return "rand", nil
}
