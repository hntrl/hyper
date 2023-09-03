package database

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type Filter interface {
	Validate(Model) error
}

type ScalarCondition int64

const (
	_ ScalarCondition = iota
	ScalarConditionEquals
	ScalarConditionContains
	ScalarConditionStartsWith
	ScalarConditionEndsWith
	ScalarConditionLessThan
	ScalarConditionLessThanEquals
	ScalarConditionGreaterThan
	ScalarConditionGreaterThanEquals
	ScalarConditionIn
	ScalarConditionIsSet
)

type ScalarFilter struct {
	Condition ScalarCondition
	Mode      QueryMode
	Key       FieldName
	Value     interface{}
}

func filterError_ForbiddenOptionalCondition(key FieldName) error {
	return fmt.Errorf("field %s filter condition %v not allowed for non-optional field", key, ScalarConditionIsSet)
}
func filterError_InvalidCondition(key FieldName) error {
	return fmt.Errorf("field %s has invalid filter condition", key)
}

func scalarFilterError(filter ScalarFilter, err error) error {
	return errors.Wrapf(err, "field %s filter value for condition %v", filter.Key, filter.Condition)
}
func scalarFilterError_ForbiddenCondition(key FieldName, condition ScalarCondition, fieldType ScalarType) error {
	return fmt.Errorf("field %s filter condition %v not allowed for field type %s", key, condition, fieldType)
}

func (filter ScalarFilter) Validate(m Model) error {
	field, err := m.GetField(filter.Key)
	if err != nil {
		return err
	}
	scalarType, ok := field.Type.(ScalarType)
	if !ok {
		return fmt.Errorf("field %s has invalid scalar type", filter.Key)
	}
	switch scalarType {
	case ScalarTypeString:
		switch filter.Condition {
		case ScalarConditionEquals, ScalarConditionContains, ScalarConditionStartsWith, ScalarConditionEndsWith:
			if err := scalarType.ValidateInterface(filter.Value); err != nil {
				return scalarFilterError(filter, err)
			}
			return nil
		case ScalarConditionLessThan, ScalarConditionLessThanEquals, ScalarConditionGreaterThan, ScalarConditionGreaterThanEquals:
			return scalarFilterError_ForbiddenCondition(filter.Key, filter.Condition, scalarType)
		case ScalarConditionIn:
			if _, ok := filter.Value.([]string); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected []string, got %T", filter.Value))
			}
			return nil
		case ScalarConditionIsSet:
			if !field.Optional {
				return filterError_ForbiddenOptionalCondition(filter.Key)
			}
			if _, ok := filter.Value.(bool); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
			}
			return nil
		default:
			return filterError_InvalidCondition(filter.Key)
		}
	case ScalarTypeInt, ScalarTypeBigInt:
		switch filter.Condition {
		case ScalarConditionEquals, ScalarConditionLessThan, ScalarConditionLessThanEquals, ScalarConditionGreaterThan, ScalarConditionGreaterThanEquals:
			if err := scalarType.ValidateInterface(filter.Value); err != nil {
				return scalarFilterError(filter, err)
			}
			return nil
		case ScalarConditionContains, ScalarConditionStartsWith, ScalarConditionEndsWith:
			return scalarFilterError_ForbiddenCondition(filter.Key, filter.Condition, scalarType)
		case ScalarConditionIn:
			if _, ok := filter.Value.([]int64); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected []int64, got %T", filter.Value))
			}
			return nil
		case ScalarConditionIsSet:
			if !field.Optional {
				return filterError_ForbiddenOptionalCondition(filter.Key)
			}
			if _, ok := filter.Value.(bool); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
			}
			return nil
		default:
			return filterError_InvalidCondition(filter.Key)
		}
	case ScalarTypeFloat, ScalarTypeDecimal:
		switch filter.Condition {
		case ScalarConditionEquals, ScalarConditionLessThan, ScalarConditionLessThanEquals, ScalarConditionGreaterThan, ScalarConditionGreaterThanEquals:
			if err := scalarType.ValidateInterface(filter.Value); err != nil {
				return scalarFilterError(filter, err)
			}
			return nil
		case ScalarConditionContains, ScalarConditionStartsWith, ScalarConditionEndsWith:
			return scalarFilterError_ForbiddenCondition(filter.Key, filter.Condition, scalarType)
		case ScalarConditionIn:
			if _, ok := filter.Value.([]float64); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected []float64, got %T", filter.Value))
			}
			return nil
		case ScalarConditionIsSet:
			if !field.Optional {
				return filterError_ForbiddenOptionalCondition(filter.Key)
			}
			if _, ok := filter.Value.(bool); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
			}
			return nil
		default:
			return filterError_InvalidCondition(filter.Key)
		}
	case ScalarTypeBoolean:
		switch filter.Condition {
		case ScalarConditionEquals:
			if err := scalarType.ValidateInterface(filter.Value); err != nil {
				return scalarFilterError(filter, err)
			}
			return nil
		case ScalarConditionContains, ScalarConditionStartsWith, ScalarConditionEndsWith, ScalarConditionLessThan, ScalarConditionLessThanEquals, ScalarConditionGreaterThan, ScalarConditionGreaterThanEquals, ScalarConditionIn:
			return scalarFilterError_ForbiddenCondition(filter.Key, filter.Condition, scalarType)
		case ScalarConditionIsSet:
			if !field.Optional {
				return filterError_ForbiddenOptionalCondition(filter.Key)
			}
			if _, ok := filter.Value.(bool); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
			}
			return nil
		default:
			return filterError_InvalidCondition(filter.Key)
		}
	case ScalarTypeDateTime:
		switch filter.Condition {
		case ScalarConditionEquals, ScalarConditionLessThan, ScalarConditionLessThanEquals, ScalarConditionGreaterThan, ScalarConditionGreaterThanEquals:
			if err := scalarType.ValidateInterface(filter.Value); err != nil {
				return scalarFilterError(filter, err)
			}
			return nil
		case ScalarConditionContains, ScalarConditionStartsWith, ScalarConditionEndsWith:
			return scalarFilterError_ForbiddenCondition(filter.Key, filter.Condition, scalarType)
		case ScalarConditionIn:
			if _, ok := filter.Value.([]time.Time); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected []time.Time, got %T", filter.Value))
			}
			return nil
		case ScalarConditionIsSet:
			if !field.Optional {
				return filterError_ForbiddenOptionalCondition(filter.Key)
			}
			if _, ok := filter.Value.(bool); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
			}
			return nil
		default:
			return filterError_InvalidCondition(filter.Key)
		}
	case ScalarTypeBytes:
		switch filter.Condition {
		case ScalarConditionEquals, ScalarConditionContains, ScalarConditionStartsWith, ScalarConditionEndsWith:
			if err := scalarType.ValidateInterface(filter.Value); err != nil {
				return scalarFilterError(filter, err)
			}
			return nil
		case ScalarConditionLessThan, ScalarConditionLessThanEquals, ScalarConditionGreaterThan, ScalarConditionGreaterThanEquals:
			return scalarFilterError_ForbiddenCondition(filter.Key, filter.Condition, scalarType)
		case ScalarConditionIn:
			if _, ok := filter.Value.([][]byte); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected [][]byte, got %T", filter.Value))
			}
			return nil
		case ScalarConditionIsSet:
			if !field.Optional {
				return filterError_ForbiddenOptionalCondition(filter.Key)
			}
			if _, ok := filter.Value.(bool); !ok {
				return scalarFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
			}
			return nil
		default:
			return filterError_InvalidCondition(filter.Key)
		}
	default:
		return fmt.Errorf("field %s filter value has an invalid data type", filter.Key)
	}
}

type EnumCondition int64

const (
	_ EnumCondition = iota
	EnumConditionEquals
	EnumConditionIn
	EnumConditionIsSet
)

type EnumFilter struct {
	Condition EnumCondition
	Key       FieldName
	Value     interface{}
}

func enumFilterError(filter EnumFilter, err error) error {
	return errors.Wrapf(err, "field %s filter value for condition %v", filter.Key, filter.Condition)
}

func (filter EnumFilter) Validate(m Model) error {
	field, err := m.GetField(filter.Key)
	if err != nil {
		return err
	}
	enumType, ok := field.Type.(EnumType)
	if !ok {
		return fmt.Errorf("cannot use enum filter on field %s with non-enum type %s", filter.Key, field.Type.TypeName())
	}
	switch filter.Condition {
	case EnumConditionEquals:
		if err := enumType.ValidateInterface(filter.Value); err != nil {
			return enumFilterError(filter, err)
		}
		return nil
	case EnumConditionIn:
		values, ok := filter.Value.([]string)
		if !ok {
			return enumFilterError(filter, fmt.Errorf("expected []string, got %T", filter.Value))
		}
		for _, value := range values {
			if err := enumType.ValidateInterface(value); err != nil {
				return enumFilterError(filter, err)
			}
		}
		return nil
	case EnumConditionIsSet:
		if !field.Optional {
			return filterError_ForbiddenOptionalCondition(filter.Key)
		}
		if _, ok := filter.Value.(bool); !ok {
			return enumFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
		}
		return nil
	default:
		return filterError_InvalidCondition(filter.Key)
	}
}

type ListCondition int64

const (
	_ ListCondition = iota
	ListConditionEquals
	ListConditionHas
	ListConditionHasEvery
	ListConditionHasSome
	ListConditionIsEmpty
	ListConditionIsSet
)

type ListFilter struct {
	Condition ListCondition
	Key       FieldName
	Value     interface{}
}

func listFilterError(filter ListFilter, err error) error {
	return errors.Wrapf(err, "field %s filter value for condition %v", filter.Key, filter.Condition)
}

func (filter ListFilter) Validate(m Model) error {
	field, err := m.GetField(filter.Key)
	if err != nil {
		return err
	}
	listType, ok := field.Type.(ListType)
	if !ok {
		return fmt.Errorf("cannot use list filter on field %s with non-list type %s", filter.Key, field.Type.TypeName())
	}
	switch filter.Condition {
	case ListConditionEquals, ListConditionHasEvery, ListConditionHasSome:
		if err := listType.ValidateInterface(filter.Value); err != nil {
			return listFilterError(filter, err)
		}
		return nil
	case ListConditionHas:
		if err := listType.ItemType.ValidateInterface(filter.Value); err != nil {
			return listFilterError(filter, err)
		}
		return nil
	case ListConditionIsEmpty:
		if _, ok := filter.Value.(bool); !ok {
			return listFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
		}
		return nil
	case ListConditionIsSet:
		if !field.Optional {
			return filterError_ForbiddenOptionalCondition(filter.Key)
		}
		if _, ok := filter.Value.(bool); !ok {
			return listFilterError(filter, fmt.Errorf("expected bool, got %T", filter.Value))
		}
		return nil
	default:
		return filterError_InvalidCondition(filter.Key)
	}
}

type CompositeCondition int64

const (
	_ CompositeCondition = iota
	CompositeConditionAND
	CompositeConditionOR
	CompositeConditionNOT
)

type CompositeFilter struct {
	Condition CompositeCondition
	Filters   []Filter
}

func (filter CompositeFilter) Validate(m Model) error {
	for _, subFilter := range filter.Filters {
		if err := subFilter.Validate(m); err != nil {
			return err
		}
	}
	return nil
}
