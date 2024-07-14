package main

import "errors"

func (b bounds) bound(toBound any) (any, error) {
	minimal := b.Min
	maximal := b.Max

	var floatVal float64
	var intVal int64
	var ok bool

	intVal, ok = toBound.(int64)
	if ok {
		if intVal < int64(minimal) {
			return minimal, nil
		}

		if intVal > int64(maximal) {
			return int(maximal), nil
		}

		return intVal, nil
	}

	floatVal, ok = toBound.(float64)
	if !ok {
		return nil, errors.New("value to bound is not a number")
	}

	if floatVal < minimal {
		return minimal, nil
	}

	if floatVal > maximal {
		return maximal, nil
	}

	return floatVal, nil
}
