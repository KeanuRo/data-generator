package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

const TYPE_ATTRIBUTE = "attribute"
const TYPE_SYSLOG = "syslog"
const TYPE_NETFLOW = "netflow"
const RAND = "rand"
const FROM_LIST = "from-list"

type initialBounds struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type Options struct {
	InitialBounds initialBounds  `json:"initial_bounds"`
	Post          string         `json:"post"`
	List          map[string]any `json:"list"`
	PickMode      string         `json:"pick-mode"`
}

type Variable struct {
	Type    string  `json:"type"`
	Options Options `json:"opts"`
}

type Plugin interface {
	calculate(options Options, cacheI cacheItem) (any, *cacheItem, error)
}

func (variable Variable) getPlugin() (Plugin, error) {
	switch variable.Type {
	case RAND:
		return randNumber{}, nil
	case FROM_LIST:
		return list{}, nil
	}

	return nil, errors.New("plugin type not supported")
}

type GenerationRule struct {
	Ident      string              `json:"ident"`
	Type       string              `json:"type"`
	Variables  map[string]Variable `json:"vars"`
	Iterations int                 `json:"iters"`
	//AllUnique  bool                `json:"all-uniq"`
	Combine bool `json:"combine"`
	Format  any  `json:"fmt"`
}

type GenerationRules []GenerationRule

type GeneratorObject struct {
	Id              int
	Schedule        string
	GenerationRules GenerationRules
}

type GeneratorObjects []GeneratorObject

type Result struct {
	linkedId      int
	generatorId   int
	Type          string
	AbstractValue string
	Attribute     Attribute
	Cache         []cacheItem
}

func (generatorObjects GeneratorObjects) Calculate(linkedObj *LinkedObject, ch chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, generatorObject := range generatorObjects {
	loop:
		for _, generationRule := range generatorObject.GenerationRules {
			calculatedIterations := make([]string, 0)
			accumulatedCache := make([]cacheItem, 0)
			for count := range generationRule.Iterations {
				calculatedVariables := make(map[string]any)
				for variableName, ruleVar := range generationRule.Variables {
					plugin, err := ruleVar.getPlugin()
					if err != nil {
						break loop
					}

					res, cacheI, err := plugin.calculate(ruleVar.Options, varCache.get(linkedObj.ID, generatorObject.Id, count, variableName))
					if err != nil {
						break loop
					}
					if cacheI != nil {
						cacheI.variable = variableName
						cacheI.iteration = count
						accumulatedCache = append(accumulatedCache, *cacheI)
					}

					calculatedVariables[variableName] = res
				}

				format, err := applyFormat(generationRule.Format, calculatedVariables, linkedObj.Attributes)
				if err != nil {
					break loop
				}

				calculatedIterations = append(calculatedIterations, format)
			}

			var value string
			result := Result{linkedId: linkedObj.ID, generatorId: generatorObject.Id, Type: generationRule.Type, Cache: accumulatedCache}
			switch generationRule.Type {
			case TYPE_ATTRIBUTE:
				if generationRule.Combine {
					raw, err := json.Marshal(calculatedIterations)
					if err != nil {
						break loop
					}
					value = string(raw)
				} else {
					value = calculatedIterations[0]
				}

				attribute, ok := linkedObj.Attributes[generationRule.Ident]
				if !ok {
					break loop
				}

				attribute.Value = value
				linkedObj.Attributes[generationRule.Ident] = attribute

				result.AbstractValue = value
				result.Attribute = attribute

				ch <- result
			case TYPE_NETFLOW, TYPE_SYSLOG:
				break loop
			default:
				break loop
			}
		}
	}
}

func applyFormat(format any, calculatedVars map[string]any, attributes map[string]Attribute) (string, error) {
	readyFormat, ok := format.(string)
	if !ok {
		raw, err := json.Marshal(format)
		if err != nil {
			return "", err
		}
		readyFormat = string(raw[:])
	}

	for key, value := range calculatedVars {
		var stringValue string
		switch value.(type) {
		case string:
			stringValue = value.(string)
		case float64, float32:
			stringValue = fmt.Sprintf("%f", value)
		case int64, int:
			stringValue = fmt.Sprintf("%d", value)
		default:
			return "", errors.New("invalid format")
		}
		readyFormat = strings.ReplaceAll(readyFormat, "%"+key+"%", stringValue)
	}

	for key, attribute := range attributes {
		readyFormat = strings.ReplaceAll(readyFormat, "%"+key+"%", attribute.Value)
	}

	return readyFormat, nil
}
