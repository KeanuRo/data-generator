package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
)

const TYPE_ATTRIBUTE = "attribute"
const TYPE_SYSLOG = "syslog"
const TYPE_NETFLOW = "netflow"
const RAND = "rand"
const FROM_LIST = "from-list"

type Options struct {
	InitialBounds map[string]float32 `json:"initial_bounds"`
	Post          string             `json:"post"`
	List          map[string]any     `json:"list"`
	PickMode      string             `json:"pick-mode"`
}

type Variable struct {
	Type    string  `json:"type"`
	Options Options `json:"opts"`
}

type Plugin interface {
	calculate(options Options, linkedId, generatorId, iteration int, variable string) (any, error)
}

func (variable Variable) getPlugin() (Plugin, error) {
	switch variable.Type {
	case RAND:
		return rand{}, nil
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
	Type      string
	Result    string
	Attribute Attribute
}

func (generatorObjects GeneratorObjects) Calculate(linkedObj LinkedObject, ch chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, generatorObject := range generatorObjects {
		for _, generationRule := range generatorObject.GenerationRules {
			calculatedIterations := make([]string, 0)
		loop:
			for count := range generationRule.Iterations {
				calculatedVariables := make(map[string]any)
				for variableName, ruleVar := range generationRule.Variables {
					plugin, err := ruleVar.getPlugin()
					if err != nil {
						break loop
					}

					res, err := plugin.calculate(ruleVar.Options, linkedObj.ID, generatorObject.Id, count, variableName)
					if err != nil {
						break loop
					}

					calculatedVariables[variableName] = res
				}

				formatType := reflect.TypeOf(generationRule.Format).Kind()
				var format string
				if formatType == reflect.String {
					format = generationRule.Format.(string)
				} else {
					raw, err := json.Marshal(generationRule.Format)
					format = string(raw[:])
					if err != nil {
						break loop
					}
				}

				calculatedIterations = append(calculatedIterations, format)
			}

			attribute, ok := linkedObj.Attributes[generationRule.Ident]
			if !ok {
				attribute = Attribute{}
			}

			calculated := calculatedIterations[0]

			result := Result{Type: generationRule.Type, Result: calculated, Attribute: attribute}

			ch <- result
		}
	}
}
