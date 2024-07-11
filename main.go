package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	start := time.Now()

	db := newDB()
	defer db.Close()

	reader := Reader{db: db}
	objects := reader.GetObjectsToProcess()

	varCache = Cacher{}
	err := varCache.load()
	if err != nil {
		panic(err)
	}

	channel := make(chan Result)
	wg := new(sync.WaitGroup)

	for _, object := range objects {
		wg.Add(1)
		go object.generatorObjects.Calculate(object, channel, wg)
	}

	results := make([]Result, 0)
	go func() {
		for result := range channel {
			results = append(results, result)
		}
	}()

	wg.Wait()
	close(channel)
	varCache.flush()

	writer := Writer{}
	writer.initialize()
	writer.SetTime(time.Now())

	for _, result := range results {
		writer.Remember(result)
		cacheItems := result.Cache
		for _, cacheI := range cacheItems {
			varCache.write(result.linkedId, result.generatorId, cacheI)
		}
	}

	writer.Prepare()

	err = varCache.save()
	if err != nil {
		panic(err)
	}

	fmt.Println(time.Now().Sub(start))
}
