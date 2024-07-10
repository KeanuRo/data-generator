package main

import (
	"encoding/json"
	"os"
)

type varCache map[int]map[int]map[int]map[string]map[string]string

func (c *varCache) save() {
	os.MkdirAll("cache", 0777)

	file, err := os.Create("cache/cache.json")
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(*c)

	_, err = file.Write(data)
	if err != nil {
		panic(err)
	}
}

func (c *varCache) get(linkedId, generatorId, iteration int, variable, key string) string {
	result, ok := (*c)[linkedId][generatorId][iteration][variable][key]

	if ok {
		return result
	}

	return ""
}

func (c *varCache) write(linkedId, generatorId, iteration int, variable, key, value string) {
	ch := *c
	_, ok := ch[linkedId][generatorId]
	if !ok {
		ch[linkedId] = make(map[int]map[int]map[string]map[string]string)
	}

	_, ok = ch[linkedId][generatorId][iteration]
	if !ok {
		ch[linkedId][generatorId][iteration] = make(map[string]map[string]string)
	}

	_, ok = ch[linkedId][generatorId][iteration][variable]
	if !ok {
		ch[linkedId][generatorId][iteration][variable] = make(map[string]string)
	}

	ch[linkedId][generatorId][iteration][variable][key] = value

	*c = ch
}
