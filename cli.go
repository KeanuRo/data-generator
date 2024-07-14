package main

import "fmt"

func reportErrorLong(linkedId, generatorId, iteration int, varName string, err error) {
	fmt.Printf("Encountered error in linked object [%d], generator object [%d], iteration [%d], variable [%s], \n\tmessage: %s\n",
		linkedId, generatorId, iteration, varName, err.Error())
}

func reportError(linkedId, generatorId, iteration int, err error) {
	fmt.Printf("Encountered error in linked object [%d], generator object [%d], iteration [%d], \n\tmessage: %s\n",
		linkedId, generatorId, iteration, err.Error())
}

func reportErrorShort(linkedId, generatorId int, err error) {
	fmt.Printf("Encountered error in linked object [%d], generator object [%d], message: \n\t%s\n",
		linkedId, generatorId, err.Error())
}
