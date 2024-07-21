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

	writer := Writer{}
	writer.initialize(db)
	writer.SetTime(time.Now())

	newCache := Cacher{}
	newCache.data = make(map[int]map[int]map[int]map[string]cacheItem)

	go func() {
		for result := range channel {
			writer.Remember(result)
			cacheItems := result.Cache
			for _, cacheI := range cacheItems {
				newCache.write(result.linkedId, result.generatorId, cacheI)
			}
		}
	}()

	wg.Wait()
	close(channel)

	writer.Exec()

	go newCache.save()

	fmt.Println(time.Now().Sub(start))
}
