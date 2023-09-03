package database

import (
	"fmt"
	"time"
)

type FieldName string

type Field struct {
	Name     FieldName
	Type     FieldType
	Optional bool
}

type FieldType interface {
	TypeName() string
	ValidateInterface(interface{}) error
}

type ScalarType string

const (
	ScalarTypeString   ScalarType = "String"
	ScalarTypeInt      ScalarType = "Int"
	ScalarTypeBigInt   ScalarType = "BigInt"
	ScalarTypeFloat    ScalarType = "Float"
	ScalarTypeDecimal  ScalarType = "Decimal"
	ScalarTypeBoolean  ScalarType = "Boolean"
	ScalarTypeDateTime ScalarType = "DateTime"
	ScalarTypeBytes    ScalarType = "Bytes"
)

func (s ScalarType) TypeName() string {
	return string(s)
}

func (s ScalarType) ValidateInterface(value interface{}) error {
	switch s {
	case ScalarTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case ScalarTypeInt, ScalarTypeBigInt:
		if _, ok := value.(int64); !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
	case ScalarTypeFloat, ScalarTypeDecimal:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected float, got %T", value)
		}
	case ScalarTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case ScalarTypeDateTime:
		if _, ok := value.(time.Time); !ok {
			return fmt.Errorf("expected time.Time, got %T", value)
		}
	case ScalarTypeBytes:
		if _, ok := value.([]byte); !ok {
			return fmt.Errorf("expected []byte, got %T", value)
		}
	default:
		return fmt.Errorf("unknown scalar type %s", s)
	}
	return nil
}

type EnumType struct {
	Values []string
}

func (e EnumType) TypeName() string {
	return "Enum"
}

func (e EnumType) ValidateInterface(value interface{}) error {
	if _, ok := value.(string); !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	for _, enumValue := range e.Values {
		if enumValue == value {
			return nil
		}
	}
	return fmt.Errorf("enum value must be one of %s, got %s", e.Values, value)
}

type ListType struct {
	ItemType FieldType
}

func (s ListType) TypeName() string {
	return fmt.Sprintf("[]%s", s.ItemType)
}

func (s ListType) ValidateInterface(value interface{}) error {
	values, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected []%s, got %T", s.ItemType.TypeName(), value)
	}
	for _, item := range values {
		if err := s.ItemType.ValidateInterface(item); err != nil {
			return err
		}
	}
	return nil
}

type ObjectType struct {
	Fields []Field
}

func (o ObjectType) TypeName() string {
	return "Object"
}

func (o ObjectType) ValidateInterface(value interface{}) error {
	// should map values be validated/allowed? the idiomatic way to do this (right now) is to create filters for each property
	return fmt.Errorf("evaluating object fields is not supported")
}
