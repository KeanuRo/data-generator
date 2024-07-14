package main

import (
	"encoding/json"
	"os"
)

const VAR_CACHE_PATH = "/var/cache/echo-center-cliback-ml/data-generator"
const VAR_CACHE_FILE = "variables.json"

type Cacher struct {
	data map[int]map[int]map[int]map[string]cacheItem
}

type cacheItem struct {
	iteration    int
	variable     string
	CurrentValue string `json:"current_value"`
	GoReverse    bool   `json:"go_reverse"`
}

var varCache Cacher

func (c *Cacher) load() error {
	err := os.MkdirAll(VAR_CACHE_PATH, 0777)
	if err != nil {
		return err
	}

	var file *os.File
	defer file.Close()
	_, err = os.Stat(VAR_CACHE_PATH + "/" + VAR_CACHE_FILE)
	if err != nil {
		file, err = os.Create(VAR_CACHE_PATH + "/" + VAR_CACHE_FILE)
		if err != nil {
			return err
		}
		c.data = make(map[int]map[int]map[int]map[string]cacheItem)
		return nil
	}

	fileBytes, _ := os.ReadFile(VAR_CACHE_PATH + "/" + VAR_CACHE_FILE)
	if (fileBytes == nil) || (len(fileBytes) == 0) {
		c.data = make(map[int]map[int]map[int]map[string]cacheItem)
		return nil
	}

	err = json.Unmarshal(fileBytes, &c.data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cacher) save() error {
	data, err := json.Marshal(c.data)
	if err != nil {
		return err
	}

	err = os.WriteFile(VAR_CACHE_PATH+"/"+VAR_CACHE_FILE, data, 0777)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cacher) flush() {
	c.data = nil
	c.data = make(map[int]map[int]map[int]map[string]cacheItem)
}

func (c *Cacher) get(linkedId, generatorId, iteration int, variable string) cacheItem {
	result, ok := (c.data)[linkedId][generatorId][iteration][variable]

	if ok {
		return result
	}

	return cacheItem{}
}

func (c *Cacher) write(linkedId, generatorId int, value cacheItem) {
	iteration := value.iteration
	variable := value.variable
	_, ok := (c.data)[linkedId]
	if !ok {
		(c.data)[linkedId] = make(map[int]map[int]map[string]cacheItem)
	}

	_, ok = (c.data)[linkedId][generatorId]
	if !ok {
		(c.data)[linkedId][generatorId] = make(map[int]map[string]cacheItem)
	}

	_, ok = (c.data)[linkedId][generatorId][iteration]
	if !ok {
		(c.data)[linkedId][generatorId][iteration] = make(map[string]cacheItem)
	}

	(c.data)[linkedId][generatorId][iteration][variable] = value
}
