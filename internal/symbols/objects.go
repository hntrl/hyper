package symbols

import (
	"fmt"
	"math"
)

// @ 1.3.2 `Array` Object

type ArrayClass struct {
	ItemClass   Class
	descriptors *ClassDescriptors `hash:"ignore"`
}

// TODO: "cache" the creation of this array class to reduce reflection calls
func NewArrayClass(itemClass Class) ArrayClass {
	arrayClass := ArrayClass{ItemClass: itemClass}
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
