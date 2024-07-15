package main

import (
	"fmt"
	"testing"
)

func TestPostProcessing_Prepare(t *testing.T) {
	p := calculatePostProcessing("  =	SUB[~0,NRAN	D[0,500]]  ", 5)
	fmt.Println(p)
}
