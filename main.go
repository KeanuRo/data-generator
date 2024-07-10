package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	db := newDB()
	defer db.Close()

	reader := Reader{db: db}

	objects := reader.GetObjectsToProcess()

	channel := make(chan Result)
	wg := new(sync.WaitGroup)

	start := time.Now()

	for _, object := range objects {
		wg.Add(1)
		go object.generatorObjects.Calculate(*object, channel, wg)
	}

	results := make([]Result, 0)
	go func() {
		for result := range channel {
			results = append(results, result)
		}
	}()

	wg.Wait()
	close(channel)

	fmt.Println(len(results))

	fmt.Println(time.Now().Sub(start))
}
