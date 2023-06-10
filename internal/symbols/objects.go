package symbols

import (
	"fmt"
	"math"

	. "github.com/hntrl/hyper/internal/symbols/errors"
	"github.com/hntrl/hyper/internal/tokens"
)

// @ ? `Any` class

var (
	Any = AnyClass{}
)

type AnyClass struct{}

func (AnyClass) Name() string {
	return "any"
}
func (AnyClass) Descriptors() *ClassDescriptors {
	return nil
}

// @ 1.3.1 `Nilable` Object

type NilableClass struct {
	parentClass Class
	descriptors *ClassDescriptors `hash:"ignore"`
}

func NewNilableClass(parentClass Class) NilableClass {
	parentDescriptors := parentClass.Descriptors()
	nilableClass := NilableClass{parentClass: parentClass}
	nilableClass.descriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Nil, func() (*NilableValue, error) {
				return &NilableValue{
					nilableClass: nilableClass,
					setValue:     nil,
				}, nil
			}),
		},
		Comparators: ClassComparatorSet{
			Comparator(Nil, tokens.EQUALS, func(a *NilableValue) (bool, error) {
				return a.setValue == nil, nil
			}),
			Comparator(Nil, tokens.NOT_EQUALS, func(a *NilableValue) (bool, error) {
				return a.setValue != nil, nil
			}),
		},
		Properties: parentDescriptors.Properties,
	}
	if parentDescriptors.Constructors != nil {
		for hash, constructor := range parentDescriptors.Constructors {
			nilableClass.descriptors.Constructors[hash] = &ClassConstructor{
				forClass: constructor.forClass,
				handler: func(val ValueObject) (ValueObject, error) {
					constructedVal, err := constructor.handler(val)
					if err != nil {
						return nil, err
					}
					return &NilableValue{nilableClass, constructedVal}, nil
				},
			}
		}
	}
	if parentDescriptors.Operators != nil {
		parentDescriptors.Operators = ClassOperatorSet{}
		for hash, operator := range parentDescriptors.Operators {
			nilableClass.descriptors.Operators[hash] = &ClassOperator{
				operandClass: operator.operandClass,
				token:        operator.token,
				handler: func(a ValueObject, b ValueObject) (ValueObject, error) {
					nilable := a.(*NilableValue)
					if nilable.setValue == nil {
						return nil, StandardError(CannotOperateNilValue, "cannot operate on nil value")
					}
					computedVal, err := operator.handler(nilable.setValue, b)
					if err != nil {
						return nil, err
					}
					return &NilableValue{nilableClass, computedVal}, nil
				},
			}
		}
	}
	if parentDescriptors.Comparators != nil {
		for hash, comparator := range parentDescriptors.Comparators {
			nilableClass.descriptors.Comparators[hash] = &ClassComparator{
				operandClass: comparator.operandClass,
				token:        comparator.token,
				handler: func(a ValueObject, b ValueObject) (bool, error) {
					nilable := a.(*NilableValue)
					if nilable.setValue == nil {
						return false, StandardError(CannotOperateNilValue, "cannot operate on nil value")
					}
					return comparator.handler(nilable.setValue, b)
				},
			}
		}
	}
	if parentDescriptors.Enumerable != nil {
		nilableClass.descriptors.Enumerable = &ClassEnumerationRules{}
		rules := parentDescriptors.Enumerable
		if rules.GetLength != nil {
			nilableClass.descriptors.Enumerable.GetLength = func(a ValueObject) (int, error) {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return -1, StandardError(CannotEnumerateNilValue, "cannot get length of nil")
				}
				return rules.GetLength(nilable.setValue)
			}
		}
		if rules.GetIndex != nil {
			nilableClass.descriptors.Enumerable.GetIndex = func(a ValueObject, idx int) (ValueObject, error) {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return nil, StandardError(CannotEnumerateNilValue, "cannot get index of nil")
				}
				return rules.GetIndex(nilable.setValue, idx)
			}
		}
		if rules.SetIndex != nil {
			nilableClass.descriptors.Enumerable.SetIndex = func(a ValueObject, idx int, b ValueObject) error {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return StandardError(CannotEnumerateNilValue, "cannot set length of nil")
				}
				return rules.SetIndex(a, idx, b)
			}
		}
		if rules.GetRange != nil {
			nilableClass.descriptors.Enumerable.GetRange = func(a ValueObject, start int, end int) (ValueObject, error) {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return nil, StandardError(CannotEnumerateNilValue, "cannot get range of nil")
				}
				return rules.GetRange(nilable.setValue, start, end)
			}
		}
		if rules.SetRange != nil {
			nilableClass.descriptors.Enumerable.SetRange = func(a ValueObject, start int, end int, b ValueObject) error {
				nilable := a.(*NilableValue)
				if nilable.setValue == nil {
					return StandardError(CannotEnumerateNilValue, "cannot set range of nil")
				}
				return rules.SetRange(nilable.setValue, start, end, b)
			}
		}
	}
	return nilableClass
}

func (nc NilableClass) Name() string {
	return fmt.Sprintf("%s?", nc.parentClass.Name())
}

func (nc NilableClass) Descriptors() *ClassDescriptors {
	return nc.descriptors
}

type NilableValue struct {
	nilableClass NilableClass
	setValue     ValueObject
}

func NewNilableValue(parentClass Class, value ValueObject) *NilableValue {
	return &NilableValue{
		nilableClass: NewNilableClass(parentClass),
		setValue:     value,
	}
}

func (no *NilableValue) Class() Class {
	return no.nilableClass
}
func (no *NilableValue) Value() interface{} {
	if no.setValue == nil {
		return nil
	}
	return no.setValue.Value()
}
func (no *NilableValue) ValueObject() ValueObject {
	return no.setValue
}

// @ 1.3.2 `Array` Object

type ArrayClass struct {
	itemClass   Class
	descriptors *ClassDescriptors `hash:"ignore"`
}

// TODO: "cache" the creation of this array class to reduce reflection calls
func NewArrayClass(itemClass Class) ArrayClass {
	arrayClass := ArrayClass{
		itemClass: itemClass,
	}
	arrayClass.descriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(arrayClass, func(arr ArrayValue) (ArrayValue, error) {
				newArray := ArrayValue{parentClass: arr.parentClass, items: make([]ValueObject, len(arr.items))}
				copy(arr.items, newArray.items)
				return newArray, nil
			}),
		},
		Enumerable: NewClassEnumerationRules(ClassEnumerationOptions{
			GetIndex: func(arr *ArrayValue, index int) (ValueObject, error) {
				if index > len(arr.items) || index < 0 {
					return nil, StandardError(IndexOutOfRange, "index out of range")
				}
				return arr.items[index], nil
			},
			SetIndex: func(arr *ArrayValue, index int, item ValueObject) error {
				if index > len(arr.items) || index < 0 {
					return StandardError(IndexOutOfRange, "index out of range")
				}
				arr.items[index] = item
				return nil
			},
			GetRange: func(arr *ArrayValue, start int, end int) (*ArrayValue, error) {
				if start > len(arr.items) || start < 0 {
					return nil, StandardError(IndexOutOfRange, "start index out of range")
				}
				if end > len(arr.items) || end < 0 {
					return nil, StandardError(IndexOutOfRange, "end index out of range")
				}
				shouldReverseOrder := start > end
				if shouldReverseOrder {
					start := int(math.Min(float64(start), float64(end)))
					end := int(math.Max(float64(start), float64(end)))
					s := make([]ValueObject, len(arr.items))
					// thank you s/o :) https://stackoverflow.com/a/19239850
					copy(s, arr.items)
					for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
						s[i], s[j] = s[j], s[i]
					}
					return &ArrayValue{
						parentClass: arr.parentClass,
						items:       s[start:end],
					}, nil
				} else {
					return &ArrayValue{
						parentClass: arr.parentClass,
						items:       arr.items[start:end],
					}, nil
				}
			},
			SetRange: func(arr *ArrayValue, start int, end int, insertArr ArrayValue) error {
				if start > len(arr.items) || start < 0 {
					return StandardError(IndexOutOfRange, "start index out of range")
				}
				if end > len(arr.items) || end < 0 {
					return StandardError(IndexOutOfRange, "end index out of range")
				}
				if start > end {
					return StandardError(InvalidRangeIndices, "start index cannot be greater than end index")
				}
				tailItems := arr.items[end:len(arr.items)]
				arr.items = append(arr.items[0:start], insertArr.items...)
				arr.items = append(arr.items, tailItems...)
				return nil
			},
		}),
		Prototype: ClassPrototypeMap{
			"append": NewClassMethod(ClassMethodOptions{
				Class:     arrayClass,
				Arguments: []Class{arrayClass.itemClass},
				Returns:   nil,
				Handler: func(arr *ArrayValue, item ValueObject) error {
					arr.items = append(arr.items, item)
					return nil
				},
			}),
			"length": NewClassMethod(ClassMethodOptions{
				Class:     arrayClass,
				Arguments: []Class{},
				Returns:   Integer,
				Handler: func(arr *ArrayValue) (IntegerValue, error) {
					return IntegerValue(len(arr.items)), nil
				},
			}),
		},
	}
	return arrayClass
}

func (ac ArrayClass) Name() string {
	return fmt.Sprintf("[]%s", ac.itemClass.Name())
}

func (ac ArrayClass) Descriptors() *ClassDescriptors {
	return ac.descriptors
}

type ArrayValue struct {
	parentClass ArrayClass
	items       []ValueObject `hash:"ignore"`
}

func NewArray(itemClass Class, len int) *ArrayValue {
	arrayClass := NewArrayClass(itemClass)
	return &ArrayValue{
		parentClass: arrayClass,
		items:       make([]ValueObject, len),
	}
}

func (av *ArrayValue) Class() Class {
	return av.parentClass
}

func (av *ArrayValue) Value() interface{} {
	out := make([]interface{}, len(av.items))
	for i, item := range av.items {
		out[i] = item.Value()
	}
	return out
}

// @ 1.3.3 `Map` Object

var Map = MapClass{}

type MapClass struct {
	Properties map[string]Class `hash:"ignore"`
}

func (MapClass) Name() string {
	return "Map"
}
func (mc MapClass) Descriptors() *ClassDescriptors {
	propertyMap := ClassPropertyMap{}
	for key, val := range mc.Properties {
		propertyMap[key] = PropertyAttributes(PropertyAttributesOptions{
			Class: val,
			Getter: func(obj *MapValue) (ValueObject, error) {
				return obj.Get(key), nil
			},
			Setter: func(obj *MapValue, val ValueObject) error {
				obj.Set(key, val)
				return nil
			},
		})
	}
	return &ClassDescriptors{
		Properties: propertyMap,
	}
}

func NewMapClass() MapClass {
	return MapClass{
		Properties: map[string]Class{},
	}
}

type MapValue struct {
	parentClass MapClass
	data        map[string]ValueObject
}

func NewMapValue() *MapValue {
	return &MapValue{
		parentClass: NewMapClass(),
		data:        map[string]ValueObject{},
	}
}

func (mv *MapValue) Class() Class {
	return mv.parentClass
}
func (mv *MapValue) Value() interface{} {
	out := make(map[string]interface{})
	for key, value := range mv.data {
		out[key] = value.Value()
	}
	return out
}

func (mv *MapValue) Get(k string) ValueObject {
	return mv.data[k]
}
func (mv *MapValue) Set(k string, v ValueObject) {
	mv.parentClass.Properties[k] = v.Class()
	mv.data[k] = v
}

// @ 1.3.4 `Error` Object

var (
	Error            = ErrorClass{}
	ErrorDescriptors = &ClassDescriptors{
		Properties: ClassPropertyMap{
			"name": PropertyAttributes(PropertyAttributesOptions{
				Class: String,
				Getter: func(a ErrorValue) (StringValue, error) {
					return StringValue(a.Name), nil
				},
			}),
			"message": PropertyAttributes(PropertyAttributesOptions{
				Class: String,
				Getter: func(a ErrorValue) (StringValue, error) {
					return StringValue(a.Message), nil
				},
			}),
		},
	}
)

type ErrorClass struct{}

func (ErrorClass) Name() string {
	return "Error"
}
func (ErrorClass) Descriptors() *ClassDescriptors {
	return ErrorDescriptors
}

type ErrorValue struct {
	Name    string
	Message string
	Data    interface{}
}

func (ErrorValue) Class() Class {
	return Error
}
func (ev ErrorValue) Value() interface{} {
	return ev
}
func (ev ErrorValue) Error() string {
	return fmt.Sprintf("%s: %s", ev.Name, ev.Message)
}
