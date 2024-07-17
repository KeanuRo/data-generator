package main

import (
	"testing"
)

func TestPostProcessing_Prepare(t *testing.T) {
	for range 100_000 {
		_, err := Calculate("=SUM[~0,RAND[0,500]]", int64(5))
		if err != nil {
			t.Error(err)
		}
	}
}
