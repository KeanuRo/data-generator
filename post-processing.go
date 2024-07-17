package main

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
)

type operand struct {
	operation string
	arguments []*operand
	value     any
}

type PostProcessor struct {
	formula string
	tree    *operand
}

func (p *PostProcessor) prepare() error {
	formula := p.formula

	replacer := strings.NewReplacer("\r", "", "\n", "", "\t", "")
	formula = replacer.Replace(formula)
	formula = strings.TrimSpace(formula)
	if formula[:1] != "=" {
		return errors.New("formula has to start with '='")
	}
	formula = formula[1:]

	formula = strings.ToLower(formula)

	characters := strings.Split(formula, "")

	stack := make([][]*operand, 0)
	depthLevel := 0
	operation := ""
	for _, character := range characters {
		switch character {
		case "{", "[", "(":
			if depthLevel == len(stack) {
				stack = append(stack, make([]*operand, 0))
			}

			newOperand := operand{operation: operation, arguments: make([]*operand, 0)}
			stack[depthLevel] = append(stack[depthLevel], &newOperand)

			if depthLevel > 0 {
				currentFather := stack[depthLevel-1][len(stack[depthLevel-1])-1]
				currentFather.arguments = append(currentFather.arguments, &newOperand)
			}

			depthLevel++

			operation = ""
		case "}", "]", ")":
			if operation != "" {
				if depthLevel == len(stack) {
					stack = append(stack, make([]*operand, 0))
				}

				newOperand := operand{value: typeConvert(operation)}
				stack[depthLevel] = append(stack[depthLevel], &newOperand)

				if depthLevel > 0 {
					currentFather := stack[depthLevel-1][len(stack[depthLevel-1])-1]
					currentFather.arguments = append(currentFather.arguments, &newOperand)
				}
			}

			depthLevel--
			operation = ""
		case ";", ",":
			if operation != "" {
				if depthLevel == len(stack) {
					stack = append(stack, make([]*operand, 0))
				}

				newOperand := operand{value: typeConvert(operation)}
				stack[depthLevel] = append(stack[depthLevel], &newOperand)

				if depthLevel > 0 {
					currentFather := stack[depthLevel-1][len(stack[depthLevel-1])-1]
					currentFather.arguments = append(currentFather.arguments, &newOperand)
				}
			}

			operation = ""
		default:
			operation += character
		}
	}

	if len(stack) == 0 {
		return errors.New("invalid formula")
	}

	if len(stack[0]) == 0 {
		return errors.New("invalid formula")
	}

	p.tree = stack[0][0]

	return nil
}

func typeConvert(value string) any {
	val, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return val
	} else {
		val, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return val
		}
	}

	return value
}

func (p *PostProcessor) calculate(input any) (any, error) {
	value, err := p.tree.evaluate(input)
	return value, err
}

func (op *operand) evaluate(input any) (any, error) {
	if op.operation == "" {
		strVal, ok := op.value.(string)
		if !ok {
			return op.value, nil
		}

		if strVal == "~0" {
			return input, nil
		}

		return op.value, nil
	}

	rawArguments := op.arguments
	arguments := make([]any, 0)
	for _, rawArgument := range rawArguments {
		arg, err := rawArgument.evaluate(input)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, arg)
	}
	switch op.operation {
	case "rand":
		if len(arguments) != 2 {
			return nil, errors.New("not enough arguments for rand")
		}

		minimum, ok := arguments[0].(int64)
		if !ok {
			return nil, errors.New("invalid argument for rand")
		}

		maximum, ok := arguments[1].(int64)
		if !ok {
			return nil, errors.New("invalid argument for rand")
		}

		return int64(rand.Intn(int(maximum-minimum)) + int(minimum)), nil
	case "sum":
		if len(arguments) != 2 {
			return nil, errors.New("not enough arguments for rand")
		}

		intFirst, ok := arguments[0].(int64)
		if !ok {
			floatFirst, ok := arguments[0].(float64)
			if !ok {
				return nil, errors.New("invalid argument for sum")
			}

			floatSecond, ok := arguments[1].(float64)
			if !ok {
				return nil, errors.New("invalid argument for sum")
			}

			return floatFirst + floatSecond, nil
		}

		intSecond, ok := arguments[1].(int64)
		if !ok {
			return nil, errors.New("invalid argument for sum")
		}

		return intFirst + intSecond, nil
	}

	return nil, nil
}

func Calculate(formula string, value any) (any, error) {
	postProcessor := PostProcessor{formula: formula}
	err := postProcessor.prepare()
	if err != nil {
		return nil, err
	}

	result, err := postProcessor.calculate(value)
	if err != nil {
		return nil, err
	}

	return result, nil
}
