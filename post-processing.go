package main

import (
	"errors"
	"strings"
)

type operand struct {
	operation string
	arguments []int
	value     string
}

type level struct {
	operands []operand
}

type PostProcessing struct {
	formula string
	levels  []level
}

func (p *PostProcessing) Prepare() error {
	p.levels = make([]level, 0)
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

	fatherStack := make([]int, 0)
	depthLevel := 0
	operation := ""
	for _, character := range characters {
		switch character {
		case "{", "[", "(":
			if depthLevel == len(p.levels) {
				p.levels = append(p.levels, level{})
			}

			newOperand := operand{operation: operation, arguments: make([]int, 0)}
			p.levels[depthLevel].operands = append(p.levels[depthLevel].operands, newOperand)

			if depthLevel > 0 {
				currentFather := fatherStack[len(fatherStack)-1]
				p.levels[depthLevel-1].operands[currentFather].arguments =
					append(p.levels[depthLevel-1].operands[currentFather].arguments,
						len(p.levels[depthLevel].operands)-1,
					)
			}

			fatherStack = append(fatherStack, len(p.levels[depthLevel].operands)-1)

			depthLevel++

			operation = ""
		case "}", "]", ")":
			if operation != "" {
				if depthLevel == len(p.levels) {
					p.levels = append(p.levels, level{})
				}
				newOperand := operand{value: operation}
				p.levels[depthLevel].operands = append(p.levels[depthLevel].operands, newOperand)

				if depthLevel > 0 {
					currentFather := fatherStack[len(fatherStack)-1]
					p.levels[depthLevel-1].operands[currentFather].arguments =
						append(p.levels[depthLevel-1].operands[currentFather].arguments, len(p.levels[depthLevel].operands)-1)
				}
			}

			depthLevel--
			fatherStack = fatherStack[:len(fatherStack)-1]
			operation = ""
		case ";", ",":
			if operation != "" {
				if depthLevel == len(p.levels) {
					p.levels = append(p.levels, level{})
				}
				newOperand := operand{value: operation}
				p.levels[depthLevel].operands = append(p.levels[depthLevel].operands, newOperand)

				if depthLevel > 0 {
					currentFather := fatherStack[len(fatherStack)-1]
					p.levels[depthLevel-1].operands[currentFather].arguments =
						append(p.levels[depthLevel-1].operands[currentFather].arguments, len(p.levels[depthLevel].operands)-1)
				}
			}

			operation = ""
		default:
			operation += character
		}
	}

	return nil
}

func calculatePostProcessing(formula string, value any) any {
	postProcessing := PostProcessing{formula: formula}
	err := postProcessing.Prepare()
	if err != nil {
		return err
	}
	return postProcessing
}
