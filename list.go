package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
)

const (
	MODE_RAND                 = "rand"
	MODE_SEQUENCE_LOOP        = "sequence-loop"
	MODE_SEQUECE_WITH_REVERSE = "sequence-with-reverse"
	MODE_TIME                 = "time"
)

const (
	TYPE_INT    = "int"
	TYPE_FLOAT  = "float"
	TYPE_STRING = "string"
)

type list struct {
	data    map[string]any
	keys    []string
	keyType string
}

func (l list) calculate(options Options, cacheI cacheItem) (any, *cacheItem, error) {
	valuesList := options.List
	if valuesList == nil || len(valuesList) == 0 {
		return nil, nil, errors.New("no values in list")
	}

	l.data = valuesList
	for key := range valuesList {
		l.keys = append(l.keys, key)
	}
	l.keyType = l.getTypeOfKeys()

	var err error
	l.keys, err = l.sortKeys()
	if err != nil {
		return nil, nil, err
	}

	_, ok := valuesList[cacheI.CurrentValue]

	if len(valuesList) == 1 || !ok {
		newValue := l.keys[0]
		cacheI.CurrentValue = newValue
		return newValue, &cacheI, nil
	}

	var result string
	switch options.PickMode {
	case MODE_RAND:
		candidate := l.keys[rand.Intn(len(l.keys))]
		chosen, err := l.randomKeyWithProbability(cacheI.CurrentValue, candidate)
		if err != nil {
			return nil, nil, err
		}
		result = chosen
		cacheI.CurrentValue = chosen
	case MODE_SEQUENCE_LOOP:
		candidate, err := l.searchInListStandard(cacheI.CurrentValue)
		chosen, err := l.randomKeyWithProbability(cacheI.CurrentValue, candidate)
		if err != nil {
			return nil, nil, err
		}
		cacheI.CurrentValue = chosen
		result = chosen
	case MODE_SEQUECE_WITH_REVERSE:
		candidate, err := l.searchInListWithReverse(&cacheI)

		if err != nil {
			return nil, nil, err
		}

		chosen, err := l.randomKeyWithProbability(cacheI.CurrentValue, candidate)
		if err != nil {
			return nil, nil, err
		}

		if chosen == l.keys[len(l.keys)-1] {
			cacheI.GoReverse = true
		}
		if chosen == l.keys[0] {
			cacheI.GoReverse = false
		}

		result = chosen
	case MODE_TIME:
		return int64(5), &cacheI, nil
	default:
		return nil, nil, errors.New(fmt.Sprintf("mode [%s] not supported", options.PickMode))
	}

	return l.transformResult(result), &cacheI, nil
}

func (l list) randomKeyWithProbability(current, candidate string) (string, error) {
	randomValue := float64(rand.Intn(10000)) / 100
	currentProbability, ok := l.data[current].(float64)
	if !ok {
		return "", errors.New("probability has to be numeric")
	}

	if randomValue < currentProbability {
		return candidate, nil
	} else {
		return current, nil
	}
}

func (l list) searchInListStandard(current string) (string, error) {
	if current == l.keys[len(l.keys)-1] {
		return l.keys[0], nil
	}

	for i, key := range l.keys {
		candidate := key
		if candidate == current {
			return l.keys[i+1], nil
		}
	}

	return "", errors.New("unexpected behaviour in standard list")
}

func (l list) searchInListWithReverse(cacheI *cacheItem) (string, error) {
	current := cacheI.CurrentValue

	if cacheI.GoReverse {
		if current == l.keys[0] {
			return l.keys[1], nil
		}

		for i := len(l.keys) - 1; i >= 0; i-- {
			candidate := l.keys[i]
			if candidate == current {
				return l.keys[i-1], nil
			}
		}
	}

	if current == l.keys[len(l.keys)-1] {
		return l.keys[0], nil
	}

	for i, key := range l.keys {
		candidate := key
		if candidate == current {
			return l.keys[i+1], nil
		}
	}

	return "", errors.New("unexpected behaviour in reverse list")
}

func (l list) getTypeOfKeys() string {
	first := l.keys[0]

	_, err := strconv.ParseInt(first, 10, 64)
	if err == nil {
		return TYPE_INT
	}

	_, err = strconv.ParseFloat(first, 64)
	if err == nil {
		return TYPE_FLOAT
	}

	return TYPE_STRING
}

func (l list) sortKeys() ([]string, error) {
	var lastErr error

	if l.keyType == TYPE_STRING {
		sort.Strings(l.keys)
		return l.keys, nil
	}

	sort.Slice(l.keys, func(i, j int) bool {
		if l.keyType == TYPE_INT {
			a, err := strconv.ParseInt(l.keys[i], 10, 64)

			if err != nil {
				lastErr = err
				return false
			}

			b, err := strconv.ParseInt(l.keys[j], 10, 64)
			if err != nil {
				lastErr = err
				return false
			}

			return a < b

		} else {
			a, err := strconv.ParseFloat(l.keys[i], 64)

			if err != nil {
				lastErr = err
				return false
			}

			b, err := strconv.ParseFloat(l.keys[j], 64)
			if err != nil {
				lastErr = err
				return false
			}

			return a < b
		}
	})

	return l.keys, lastErr
}

func (l list) transformResult(result string) (value any) {
	switch l.keyType {
	case TYPE_INT:
		value, _ = strconv.ParseInt(result, 10, 64)
	case TYPE_FLOAT:
		value, _ = strconv.ParseFloat(result, 64)
	default:
		value = result
	}
	return value
}
