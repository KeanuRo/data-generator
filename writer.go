package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Writer struct {
	time              time.Time
	toAttributes      []string
	toHistoryTables   map[string][]string
	attributesAllowed bool
	queries           []string
	db                *sql.DB
}

func (w *Writer) initialize(db *sql.DB) {
	w.toAttributes = make([]string, 0)
	w.toHistoryTables = make(map[string][]string)
	w.attributesAllowed = true
	w.queries = make([]string, 0)
	w.db = db
}

func (w *Writer) SetTime(t time.Time) {
	w.time = t
}

func (w *Writer) setWritingToAttributes(writeToAttributes bool) {
	w.attributesAllowed = writeToAttributes
}

func (w *Writer) Remember(result Result) {
	switch result.Type {
	case TYPE_ATTRIBUTE:
		attr := result.Attribute
		if attr.HistoryToDB {
			w.toHistoryTables[attr.HistoryTable] = append(w.toHistoryTables[attr.HistoryTable],
				fmt.Sprintf("(%d, '%s', '%s')", attr.ID, w.time.Format(time.DateTime), attr.Value))
		}
		if w.attributesAllowed {
			w.toAttributes = append(w.toAttributes,
				fmt.Sprintf("(%d, '%s', '%s'::timestamp)", attr.ID, attr.Value, w.time.Format(time.DateTime)))
		}
	case TYPE_SYSLOG:
		//
	}
}

func (w *Writer) prepareToAttributes() {
	if !w.attributesAllowed {
		return
	}

	query := strings.Builder{}
	query.WriteString("UPDATE object_attributes as oa SET attribute_value = up.attribute_value, updated_at = up.updated_at FROM (VALUES")
	query.WriteString(strings.Join(w.toAttributes, ","))
	query.WriteString(") as up(id,attribute_value,updated_at) WHERE up.id = oa.id")

	w.queries = append(w.queries, query.String())
}

func (w *Writer) prepareToHistoryTables() {
	for tableName, table := range w.toHistoryTables {
		query := strings.Builder{}
		query.WriteString("INSERT INTO " + tableName + " (object_attribute_id, time,value) VALUES ")
		for i, row := range table {
			if i > 0 {
				query.WriteString(",")
			}
			query.WriteString(row)
			if i > 60_000 {
				w.queries = append(w.queries, query.String())
				query.Reset()
				query.WriteString("INSERT INTO " + tableName + " (object_attribute_id, time,value) VALUES ")
			}
		}
		w.queries = append(w.queries, query.String())
	}
}

func (w *Writer) Exec() {
	w.prepareToAttributes()
	w.prepareToHistoryTables()

	transaction, err := w.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	for _, query := range w.queries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = transaction.Exec(query)
			if err != nil {
				transaction.Rollback()
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()
	err = transaction.Commit()
	if err != nil {
		log.Fatal(err)
	}
}
