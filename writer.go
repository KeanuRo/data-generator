package main

import (
	"fmt"
	"strings"
	"time"
)

type Writer struct {
	time              time.Time
	attributeStorage  []Attribute
	cacheStorage      []Attribute
	historyStorage    HistoryTables
	syslogStorage     []string
	queries           []string
	writeToAttributes bool
}

func (w *Writer) initialize() {
	w.attributeStorage = make([]Attribute, 0)
	w.cacheStorage = make([]Attribute, 0)
	w.historyStorage = HistoryTables{}
	w.historyStorage.data = make(map[string][]Attribute)
	w.syslogStorage = make([]string, 0)
	w.queries = make([]string, 0)
	w.writeToAttributes = true
}

type HistoryTables struct {
	data map[string][]Attribute
}

func (h *HistoryTables) remember(attr Attribute) {
	_, ok := h.data[attr.HistoryTable]
	if !ok {
		h.data[attr.HistoryTable] = make([]Attribute, 0)
	}
	h.data[attr.HistoryTable] = append(h.data[attr.HistoryTable], attr)
}

func (w *Writer) SetTime(t time.Time) {
	w.time = t
}

func (w *Writer) setWritingToAttributes(writeToAttributes bool) {
	w.writeToAttributes = writeToAttributes
}

func (w *Writer) Remember(result Result) {
	switch result.Type {
	case TYPE_ATTRIBUTE:
		attr := result.Attribute
		if attr.HistoryToDB {
			w.historyStorage.remember(attr)
		}
		if attr.HistoryToCache {
			w.cacheStorage = append(w.cacheStorage, attr)
		}
		if w.writeToAttributes {
			w.attributeStorage = append(w.attributeStorage, attr)
		}
	case TYPE_SYSLOG:
		w.syslogStorage = append(w.syslogStorage, result.AbstractValue)
	}
}

func (w *Writer) Prepare() {
	w.queries = []string{}
	w.queries = append(w.queries, w.prepareAttributes())
}

func (w *Writer) prepareAttributes() string {
	query := strings.Builder{}
	query.WriteString("UPDATE object_attributes as oa SET attribute_value = up.attribute_value, updated_at = up.updated_at FROM (VALUES")
	rows := make([]string, 0)
	for _, attr := range w.attributeStorage {
		rows = append(rows, fmt.Sprintf("(%d,'%s','%s')", attr.ID, attr.Value, w.time.Format(time.DateTime)))
	}
	query.WriteString(strings.Join(rows, ","))
	query.WriteString(") as up(id,attribute_value,updated_at) WHERE up.id = tn.id")
	return query.String()
}
