package database

import (
	"fmt"
	"strings"
)

type DatabaseReader interface {
	GetRecord(Model, Filter) (Record, error)
	GetManyRecords(Model, QueryArguments) (Cursor, error)
}

type DatabaseWriter interface {
	CreateRecord(Model, WriteArguments) (Record, error)
	CreateManyRecords(Model, []WriteArguments) (uint64, error)
	UpdateOne(Model, Filter, WriteArguments) (Record, error)
	UpdateMany(Model, Filter, WriteArguments) (uint64, error)
	UpdateRecord(Model, Record, WriteArguments) (Record, error)
	DeleteOne(Model, Filter) error
	DeleteMany(Model, Filter) (uint64, error)
	DeleteRecord(Model, Record) error
}

type Connection interface {
	DatabaseReader
	DatabaseWriter
}

type Cursor interface {
	All() ([]Record, error)
	Next() bool
	Current() (*Record, error)
	RemainingLength() uint64
}

const RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"

// SortOrder describes the order of sorting
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// QueryMode describes the mode of querying
type QueryMode string

const (
	QueryModeDefault     QueryMode = "default"
	QueryModeInsensitive QueryMode = "insensitive"
)

type Model struct {
	Name   string
	Fields []Field
}

func (m Model) GetField(name FieldName) (*Field, error) {
	flatFields := m.normalizedFields()
	for _, field := range flatFields {
		if field.Name == name {
			return &field, nil
		}
	}
	return nil, fmt.Errorf("field %s does not exist on model %s", name, m.Name)
}

func (m Model) TableName() string {
	return strings.Replace(m.Name, ".", "_", -1)
}

func (m Model) ValidateFilter(filter Filter) error {
	return filter.Validate(m)
}

func normalizedFieldList(list []Field) []Field {
	out := make([]Field, 0)
	for _, field := range list {
		switch t := field.Type.(type) {
		case ObjectType:
			for _, innerField := range normalizedFieldList(t.Fields) {
				out = append(out, Field{
					Name:     FieldName(fmt.Sprintf("%s.%s", field.Name, innerField.Name)),
					Type:     innerField.Type,
					Optional: field.Optional || innerField.Optional,
				})
			}
		default:
			out = append(out, field)
		}
	}
	return out
}

// FIXME: this concept of "normalized fields" should be more tightly coupled
// with "FieldValues" (they have the same paradigm, but aren't really correlated
// in the code)
func (m Model) normalizedFields() []Field {
	return normalizedFieldList(m.Fields)
}

type WriteArguments FieldValues

func (m Model) ValidateWriteArguments(args WriteArguments) error {
	for _, field := range m.normalizedFields() {
		targetArgument, ok := args[field.Name]
		if !ok {
			if !field.Optional {
				return fmt.Errorf("field %s is required", field.Name)
			}
			continue
		}
		if err := field.Type.ValidateInterface(targetArgument); err != nil {
			return err
		}
	}
	return nil
}

type QueryArguments struct {
	Limit    uint
	Skip     uint
	Filter   Filter
	OrderBy  map[FieldName]SortOrder
	Distinct []FieldName
	GroupBy  []FieldName
}

func (m Model) ValidateQueryArguments(args QueryArguments) error {
	if err := args.Filter.Validate(m); err != nil {
		return err
	}
	for fieldName := range args.OrderBy {
		if _, err := m.GetField(fieldName); err != nil {
			return err
		}
	}
	for _, fieldName := range args.Distinct {
		if _, err := m.GetField(fieldName); err != nil {
			return err
		}
	}
	for _, fieldName := range args.GroupBy {
		if _, err := m.GetField(fieldName); err != nil {
			return err
		}
	}
	return nil
}

type Record struct {
	_id    string
	model  Model
	values FieldValues
}

// NewRecordValue acts as the constructor for Record. It should be noted that
// this doesn't imply the creation of a record, but rather the value it occupies
// in the process.
func NewRecordValue(m Model, id string, values map[string]interface{}) Record {
	return Record{
		_id:    id,
		model:  m,
		values: NewFieldValuesFromEmbeddedMap(values),
	}
}

func (r Record) Identifier() string {
	return r._id
}
func (r Record) Model() Model {
	return r.model
}
func (r Record) Values() FieldValues {
	return r.values
}
