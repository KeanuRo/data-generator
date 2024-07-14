package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

const MODE_RAND = "rand"
const MODE_SEQUENCE_LOOP = "sequence-loop"
const MODE_SEQUECE_WITH_REVERSE = "sequence-with-reverse"
const MODE_TIME = "time"

type list struct {
	data map[string]any
	keys []string
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

	_, ok := valuesList[cacheI.CurrentValue]

	if len(valuesList) == 1 || !ok {
		newValue := l.keys[0]
		cacheI.CurrentValue = newValue
		return newValue, &cacheI, nil
	}

	sort.Strings(l.keys)

	switch options.PickMode {
	case MODE_RAND:
		candidate := l.keys[rand.Intn(len(l.keys))]
		chosen, err := l.randomKeyWithProbability(cacheI.CurrentValue, candidate)
		if err != nil {
			return nil, nil, err
		}
		cacheI.CurrentValue = chosen
		return chosen, &cacheI, nil
	case MODE_SEQUENCE_LOOP:
		candidate, err := l.searchInListStandard(cacheI.CurrentValue)
		chosen, err := l.randomKeyWithProbability(cacheI.CurrentValue, candidate)
		if err != nil {
			return nil, nil, err
		}
		cacheI.CurrentValue = chosen
		return chosen, &cacheI, nil
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

		return chosen, &cacheI, nil
	case MODE_TIME:
		return 5, &cacheI, nil
	default:
		return nil, nil, errors.New(fmt.Sprintf("mode [%s] not supported", options.PickMode))
	}
}

func (l list) randomKeyWithProbability(current, candidate string) (string, error) {
	randomValue := int(rand.Intn(10000) / 100)
	currentProbability, ok := l.data[current].(int)
	if !ok {
		return "", errors.New("probability has to be type of integer")
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

	for i, key := range l.keys {
		candidate := key
		if candidate == current {
			return l.keys[i+1], nil
		}
	}

	return "", errors.New("unexpected behaviour in reverse list")
}
