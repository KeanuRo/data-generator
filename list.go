package main

type list struct{}

func (l list) calculate(options Options, cacheI cacheItem) (any, *cacheItem, error) {
	return 5, &cacheI, nil
}
