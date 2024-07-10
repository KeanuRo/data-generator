package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Attribute struct {
	ID             int
	Value          string
	HistoryTable   string
	HistoryToDB    bool
	HistoryToCache bool
}

type LinkedObject struct {
	ID               int
	Name             string
	Codename         string
	Attributes       map[string]Attribute
	generatorObjects GeneratorObjects
}

var generatorObjectsMap map[int]*GeneratorObject

var linkedObjects map[int]*LinkedObject

type Reader struct {
	db *sql.DB
}

var cache varCache

func (r *Reader) findGeneratorObjects() []string {
	rows, err := r.db.Query(
		"SELECT objects.id FROM objects " +
			"JOIN classes ON objects.class_id = classes.id " +
			"JOIN class_stereotypes ON classes.class_stereotype_id = class_stereotypes.id " +
			"WHERE class_stereotypes.mnemo = 'demo_generation_template'",
	)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	generatorObjectIds := make([]string, 0)
	generatorObjectsMap = make(map[int]*GeneratorObject)
	for rows.Next() {
		var generatorObject GeneratorObject
		rows.Scan(&generatorObject.Id)

		generatorObjectsMap[generatorObject.Id] = &generatorObject
		generatorObjectIds = append(generatorObjectIds, strconv.Itoa(generatorObject.Id))
	}

	if len(generatorObjectIds) == 0 {
		fmt.Println("No generator objects found")
		os.Exit(0)
	}

	return generatorObjectIds
}

func (r *Reader) findAGeneratorAttributes(generatorObjectIds []string) {
	rows, err := r.db.Query(
		"SELECT object_id, codename, attribute_value FROM object_attributes " +
			"JOIN attributes ON object_attributes.attribute_id = attributes.id " +
			"WHERE object_id IN (" + strings.Join(generatorObjectIds, ", ") + ") " +
			"AND attributes.codename in ('sched', 'generation_rules')")

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id int
		var codename string
		var attributeValue string

		rows.Scan(&id, &codename, &attributeValue)

		if codename == "sched" {
			generatorObjectsMap[id].Schedule = attributeValue
		} else if codename == "generation_rules" {
			generationRules := GenerationRules{}

			err = json.Unmarshal([]byte(attributeValue), &generationRules)

			if err != nil {
				continue
			}

			generatorObjectsMap[id].GenerationRules = generationRules
		}
	}
}

func (r *Reader) findLinkedObjects(generatorObjectIds []string) {
	rows, err := r.db.Query(
		"SELECT left_object_id, right_object_id, objects.name as object_name, objects.codename as object_codename FROM links " +
			"JOIN objects ON links.left_object_id = objects.id " +
			"JOIN relations ON links.relation_id = relations.id " +
			"JOIN relation_stereotypes ON relations.relation_stereotype_id = relation_stereotypes.id " +
			"WHERE mnemo = 'demo_generation_template' AND right_object_id IN (" + strings.Join(generatorObjectIds, ", ") + ")",
	)

	if err != nil {
		panic(err)
	}

	objectIds := make([]string, 0)
	linkedObjects = make(map[int]*LinkedObject)
	for rows.Next() {
		var generatorId int
		var linkedId int
		var name string
		var codename string
		rows.Scan(&linkedId, &generatorId, &name, &codename)

		generatorObject := *generatorObjectsMap[generatorId]

		linkedObject, ok := linkedObjects[linkedId]
		if !ok {
			linkedObject = &LinkedObject{ID: linkedId, Name: name, Codename: codename}
			objectIds = append(objectIds, strconv.Itoa(linkedId))
		}

		linkedObject.Attributes = make(map[string]Attribute)
		linkedObject.generatorObjects = append(linkedObject.generatorObjects, generatorObject)
		linkedObjects[linkedId] = linkedObject
	}

	rows.Close()

	if len(objectIds) == 0 {
		fmt.Println("No linked objects found")
		os.Exit(0)
	}

	rows, err = r.db.Query(
		`
			SELECT objects.id as object_id, attributes.codename, object_attributes.id as object_attribute_id, attribute_value,
				attributes.history_to_db, attributes.history_to_cache, data_types.history_table
			FROM objects 
			JOIN object_attributes ON objects.id = object_attributes.object_id
			JOIN attributes ON object_attributes.attribute_id = attributes.id
			JOIN data_types on attributes.data_type_id = data_types.id 
			` + "WHERE objects.id IN (" + strings.Join(objectIds, ", ") + ")",
	)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var objectId int
		var codename string
		var attributeId int
		var attributeValue string
		var historyToDB bool
		var historyToCache bool
		var historyTable string

		rows.Scan(&objectId, &codename, &attributeId, &attributeValue, &historyToDB, &historyToCache, &historyTable)

		attribute := Attribute{ID: attributeId, Value: attributeValue, HistoryToDB: historyToDB, HistoryToCache: historyToCache, HistoryTable: historyTable}

		linkedObjects[objectId].Attributes[codename] = attribute
	}

	generatorObjectIds = nil
	generatorObjectsMap = nil

	rows.Close()
}

func (r *Reader) GetObjectsToProcess() map[int]*LinkedObject {
	generatorObjectIds := r.findGeneratorObjects()
	r.findAGeneratorAttributes(generatorObjectIds)
	r.findLinkedObjects(generatorObjectIds)

	return linkedObjects
}
