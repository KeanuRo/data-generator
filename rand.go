package main

type randNumber struct{}

func (r randNumber) calculate(options Options, cacheI cacheItem) (any, *cacheItem, error) {
	//miN := options.InitialBounds.Min
	//maX := options.InitialBounds.Max
	//return rand.Intn(maX-miN) + miN, nil, nil

	return 5, nil, nil
}
