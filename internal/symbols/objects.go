package symbols

import (
	"fmt"
	"math"

	"github.com/hntrl/hyper/internal/tokens"
)

// @ 1.3.1 `Nilable` Object

type NilableClass struct {
	ParentClass Class
	descriptors *ClassDescriptors `hash:"ignore"`
}

func NewNilableClass(parentClass Class) NilableClass {
	parentDescriptors := parentClass.Descriptors()
	nilableClass := NilableClass{ParentClass: parentClass}
	nilableClass.descriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(Nil, func() (*NilableObject, error) {
				return &NilableObject{
					ObjectClass: nilableClass,
					Object:      nil,
				}, nil
			}),
		},
		Comparators: ClassComparatorSet{
			Comparator(Nil, tokens.EQUALS, func(a *NilableObject) (bool, error) {
				return a.Value == nil, nil
			}),
			Comparator(Nil, tokens.NOT_EQUALS, func(a *NilableObject) (bool, error) {
				return a.Value != nil, nil
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
					return &NilableObject{nilableClass, constructedVal}, nil
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
					nilable := a.(*NilableObject)
					if nilable.Object == nil {
						return nil, CannotOperateNilValueError()
					}
					computedVal, err := operator.handler(nilable.Object, b)
					if err != nil {
						return nil, err
					}
					return &NilableObject{nilableClass, computedVal}, nil
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
					nilable := a.(*NilableObject)
					if nilable.Object == nil {
						return false, CannotOperateNilValueError()
					}
					return comparator.handler(nilable.Object, b)
				},
			}
		}
	}
	if parentDescriptors.Enumerable != nil {
		nilableClass.descriptors.Enumerable = &ClassEnumerationRules{}
		rules := parentDescriptors.Enumerable
		if rules.GetLength != nil {
			nilableClass.descriptors.Enumerable.GetLength = func(a ValueObject) (int, error) {
				nilable := a.(*NilableObject)
				if nilable.Object == nil {
					return -1, CannotGetNilEnumerableLengthError()
				}
				return rules.GetLength(nilable.Object)
			}
		}
		if rules.GetIndex != nil {
			nilableClass.descriptors.Enumerable.GetIndex = func(a ValueObject, idx int) (ValueObject, error) {
				nilable := a.(*NilableObject)
				if nilable.Object == nil {
					return nil, CannotGetNilEnumerableLengthError()
				}
				return rules.GetIndex(nilable.Object, idx)
			}
		}
		if rules.SetIndex != nil {
			nilableClass.descriptors.Enumerable.SetIndex = func(a ValueObject, idx int, b ValueObject) error {
				nilable := a.(*NilableObject)
				if nilable.Object == nil {
					return CannotSetNilEnumerableIndexError()
				}
				return rules.SetIndex(a, idx, b)
			}
		}
		if rules.GetRange != nil {
			nilableClass.descriptors.Enumerable.GetRange = func(a ValueObject, start int, end int) (ValueObject, error) {
				nilable := a.(*NilableObject)
				if nilable.Object == nil {
					return nil, CannotGetNilEnumerableRangeError()
				}
				return rules.GetRange(nilable.Object, start, end)
			}
		}
		if rules.SetRange != nil {
			nilableClass.descriptors.Enumerable.SetRange = func(a ValueObject, start int, end int, b ValueObject) error {
				nilable := a.(*NilableObject)
				if nilable.Object == nil {
					return CannotSetNilEnumerableRangeError()
				}
				return rules.SetRange(nilable.Object, start, end, b)
			}
		}
	}
	return nilableClass
}

func (nc NilableClass) Name() string {
	return fmt.Sprintf("%s?", nc.ParentClass.Name())
}

func (nc NilableClass) Descriptors() *ClassDescriptors {
	return nc.descriptors
}

type NilableObject struct {
	ObjectClass NilableClass
	Object      ValueObject
}

func (no *NilableObject) Class() Class {
	return no.ObjectClass
}
func (no *NilableObject) Value() interface{} {
	if no.Object == nil {
		return nil
	}
	return no.Object.Value()
}

// @ 1.3.2 `Array` Object

type ArrayClass struct {
	ItemClass   Class
	descriptors *ClassDescriptors `hash:"ignore"`
}

// TODO: "cache" the creation of this array class to reduce reflection calls
func NewArrayClass(itemClass Class) ArrayClass {
	arrayClass := ArrayClass{
		ItemClass: itemClass,
	}
	arrayClass.descriptors = &ClassDescriptors{
		Constructors: ClassConstructorSet{
			Constructor(arrayClass, func(arr ArrayValue) (ArrayValue, error) {
				newArray := ArrayValue{ParentClass: arr.ParentClass, Items: make([]ValueObject, len(arr.Items))}
				copy(arr.Items, newArray.Items)
				return newArray, nil
			}),
		},
		Enumerable: NewClassEnumerationRules(ClassEnumerationOptions{
			GetIndex: func(arr *ArrayValue, index int) (ValueObject, error) {
				if index > len(arr.Items) || index < 0 {
					return nil, IndexOutOfRangeError()
				}
				return arr.Items[index], nil
			},
			SetIndex: func(arr *ArrayValue, index int, item ValueObject) error {
				if index > len(arr.Items) || index < 0 {
					return IndexOutOfRangeError()
				}
				arr.Items[index] = item
				return nil
			},
			GetRange: func(arr *ArrayValue, start int, end int) (*ArrayValue, error) {
				if start > len(arr.Items) || start < 0 {
					return nil, StartIndexOutOfRangeError()
				}
				if end > len(arr.Items) || end < 0 {
					return nil, EndIndexOutOfRangeError()
				}
				shouldReverseOrder := start > end
				if shouldReverseOrder {
					start := int(math.Min(float64(start), float64(end)))
					end := int(math.Max(float64(start), float64(end)))
					s := make([]ValueObject, len(arr.Items))
					// thank you s/o :) https://stackoverflow.com/a/19239850
					copy(s, arr.Items)
					for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
						s[i], s[j] = s[j], s[i]
					}
					return &ArrayValue{
						ParentClass: arr.ParentClass,
						Items:       s[start:end],
					}, nil
				} else {
					return &ArrayValue{
						ParentClass: arr.ParentClass,
						Items:       arr.Items[start:end],
					}, nil
				}
			},
			SetRange: func(arr *ArrayValue, start int, end int, insertArr ArrayValue) error {
				if start > len(arr.Items) || start < 0 {
					return StartIndexOutOfRangeError()
				}
				if end > len(arr.Items) || end < 0 {
					return EndIndexOutOfRangeError()
				}
				if start > end {
					return InvalidIndicesError()
				}
				tailItems := arr.Items[end:len(arr.Items)]
				arr.Items = append(arr.Items[0:start], insertArr.Items...)
				arr.Items = append(arr.Items, tailItems...)
				return nil
			},
		}),
		Prototype: ClassPrototypeMap{
			"append": NewClassMethod(ClassMethodOptions{
				Class:     arrayClass,
				Arguments: []Class{arrayClass.ItemClass},
				Returns:   nil,
				Handler: func(arr *ArrayValue, item ValueObject) error {
					arr.Items = append(arr.Items, item)
					return nil
				},
			}),
			"length": NewClassMethod(ClassMethodOptions{
				Class:     arrayClass,
				Arguments: []Class{},
				Returns:   Integer,
				Handler: func(arr *ArrayValue) (IntegerValue, error) {
					return IntegerValue(len(arr.Items)), nil
				},
			}),
		},
	}
	return arrayClass
}

func (ac ArrayClass) Name() string {
	return fmt.Sprintf("[]%s", ac.ItemClass.Name())
}

func (ac ArrayClass) Descriptors() *ClassDescriptors {
	return ac.descriptors
}

type ArrayValue struct {
	ParentClass ArrayClass
	Items       []ValueObject `hash:"ignore"`
}

func NewArray(itemClass Class, len int) *ArrayValue {
	arrayClass := NewArrayClass(itemClass)
	return &ArrayValue{
		ParentClass: arrayClass,
		Items:       make([]ValueObject, len),
	}
}

func (av *ArrayValue) Class() Class {
	return av.ParentClass
}

func (av *ArrayValue) Value() interface{} {
	out := make([]interface{}, len(av.Items))
	for i, item := range av.Items {
		out[i] = item.Value()
	}
	return out
}

// @ 1.3.3 `Map` Object

type MapClass struct {
	Properties map[string]Class
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
				return obj.Data[key], nil
			},
			Setter: func(obj *MapValue, val ValueObject) error {
				obj.ParentClass.Properties[key] = val.Class()
				obj.Data[key] = val
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
	ParentClass MapClass
	Data        map[string]ValueObject
}

func (mv *MapValue) Class() Class {
	return mv.ParentClass
}
func (mv *MapValue) Value() interface{} {
	out := make(map[string]interface{})
	for key, value := range mv.Data {
		out[key] = value.Value()
	}
	return out
}

func NewMapValue() *MapValue {
	return &MapValue{
		ParentClass: NewMapClass(),
		Data:        map[string]ValueObject{},
	}
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
