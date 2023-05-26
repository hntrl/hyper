package symbols

import (
	"fmt"

	"github.com/hntrl/hyper/internal/tokens"
)

type AnyClass struct{}

func (ac AnyClass) ClassName() string {
	return "any"
}
func (ac AnyClass) Constructors() ConstructorMap {
	return NewConstructorMap()
}
func (ac AnyClass) Get(key string) (Object, error) {
	return nil, nil
}

// MapObject represents a set of propertties without any strict bindings to
// a type
type MapObject struct {
	Properties map[string]Class       `hash:"ignore"`
	Data       map[string]ValueObject `hash:"ignore"`
}

func NewMapObject() *MapObject {
	return &MapObject{
		Properties: make(map[string]Class),
		Data:       make(map[string]ValueObject),
	}
}

func (obj MapObject) ClassName() string {
	return "MapObject"
}
func (obj MapObject) Fields() map[string]Class {
	return obj.Properties
}
func (obj MapObject) Constructors() ConstructorMap {
	return NewConstructorMap()
}

func (obj MapObject) Class() Class {
	return obj
}
func (obj MapObject) Value() interface{} {
	out := make(map[string]interface{})
	for key, obj := range obj.Data {
		out[key] = obj.Value()
	}
	return out
}
func (obj *MapObject) Set(key string, new ValueObject) error {
	if class, ok := obj.Properties[key]; ok {
		if !ClassEquals(class, new.Class()) {
			return fmt.Errorf("cannot assign %s to %s", new.Class().ClassName(), class.ClassName())
		}
	}
	obj.Data[key] = new
	return nil
}
func (obj MapObject) Get(key string) (Object, error) {
	return obj.Data[key], nil
}

// Type represents a structured list of fields that are strongly controlled
type Type struct {
	Name       string
	Private    bool
	Comment    string
	Properties map[string]Class
}

func (t Type) ClassName() string {
	return t.Name
}
func (t Type) Fields() map[string]Class {
	return t.Properties
}
func (t Type) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddGenericConstructor(t, func(data map[string]ValueObject) (ValueObject, error) {
		obj := TypeObject{t, data}
		for key, class := range t.Properties {
			if nilableProp, ok := class.(*NilableObject); ok {
				nilable := *nilableProp
				if nilableObject, ok := data[key].(*NilableObject); ok {
					nilable.Object = nilableObject.Object
				} else {
					nilable.Object = data[key]
				}
				obj.Data[key] = &nilable
			}
		}
		return &obj, nil
	})
	return csMap
}
func (t Type) Get(key string) (Object, error) {
	return nil, nil
}

func (t Type) Export() (Object, error) {
	return t, nil
}

// TypeObject represents an instance of a Type
type TypeObject struct {
	ParentType Type
	Data       map[string]ValueObject
}

func (to TypeObject) Class() Class {
	return to.ParentType
}
func (to TypeObject) Value() interface{} {
	out := make(map[string]interface{})
	for key, obj := range to.Data {
		out[key] = obj.Value()
	}
	return out
}
func (to *TypeObject) Set(key string, obj ValueObject) error {
	to.Data[key] = obj
	return nil
}
func (to TypeObject) Get(key string) (Object, error) {
	return to.Data[key], nil
}

type Iterable struct {
	ParentType Class
	Items      []ValueObject `hash:"ignore"`
}

func NewIterable(class Class, len int) Iterable {
	return Iterable{
		ParentType: class,
		Items:      make([]ValueObject, len),
	}
}

func (it Iterable) ClassName() string {
	return fmt.Sprintf("[]%s", it.ParentType.ClassName())
}
func (it Iterable) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddConstructor(it, func(obj ValueObject) (ValueObject, error) {
		iter := obj.(Iterable)
		newIter := NewIterable(it.ParentType, len(iter.Items))
		copy(newIter.Items, iter.Items)
		return obj, nil
	})
	return csMap
}

func (it Iterable) Class() Class {
	return it
}
func (it Iterable) Value() interface{} {
	out := make([]interface{}, len(it.Items))
	for i, obj := range it.Items {
		out[i] = obj.Value()
	}
	return out
}
func (it Iterable) Set(key string, obj ValueObject) error {
	return CannotSetPropertyError(key, it)
}
func (it Iterable) Get(key string) (Object, error) {
	methods := map[string]Object{
		"append": NewFunction(FunctionOptions{
			Arguments: []Class{
				it.ParentType,
			},
			Returns: it,
			Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
				return Iterable{
					ParentType: it.ParentType,
					Items:      append(it.Items, args[0]),
				}, nil
			},
		}),
	}
	return methods[key], nil
}

func (it Iterable) GetIndex(index int) (ValueObject, error) {
	if index > it.Len() || index < 0 {
		return nil, fmt.Errorf("index out of range")
	}
	return it.Items[index], nil
}
func (it Iterable) SetIndex(index int, obj ValueObject) error {
	it.Items[index] = obj
	return nil
}
func (it Iterable) Range(a, b int) (Indexable, error) {
	if a > it.Len() || b > it.Len() || a < 0 || b < 0 {
		return nil, fmt.Errorf("index out of range")
	}
	return Iterable{
		ParentType: it.ParentType,
		Items:      it.Items[a:b],
	}, nil
}
func (it Iterable) Len() int {
	return len(it.Items)
}

type PartialObject struct {
	ClassObject ObjectClass
	Object      ValueObject `hash:"ignore"`
}

func NewPartialObject(c ObjectClass) *PartialObject {
	return &PartialObject{ClassObject: c}
}

func (po PartialObject) ClassName() string {
	return fmt.Sprintf("Partial<%s>", po.ClassObject.ClassName())
}
func (po PartialObject) Fields() map[string]Class {
	out := make(map[string]Class)
	for k, v := range po.ClassObject.Fields() {
		if val, ok := v.(*NilableObject); ok {
			nilable := *val
			out[k] = &nilable
		} else {
			out[k] = &NilableObject{ClassObject: v}
		}
	}
	return out
}
func (po PartialObject) Constructors() ConstructorMap {
	return po.ClassObject.Constructors()
}

func (po PartialObject) Class() Class {
	return po
}
func (po PartialObject) Value() interface{} {
	return po.Object.Value()
}
func (po *PartialObject) Set(key string, obj ValueObject) error {
	return po.Object.Set(key, obj)
}
func (po PartialObject) Get(key string) (Object, error) {
	return po.Object.Get(key)
}

// TODO: Nilable types should not satisfy it's not nilable type. (i.e. String? should not satisfy String)

type NilableObject struct {
	ClassObject Class
	Object      ValueObject `hash:"ignore"`
}

func NewOptionalClass(c Class) *NilableObject {
	return &NilableObject{ClassObject: c}
}

func (no NilableObject) ClassName() string {
	return fmt.Sprintf("%s?", no.ClassObject.ClassName())
}
func (no NilableObject) Fields() map[string]Class {
	if objectClass, ok := no.ClassObject.(ObjectClass); ok {
		return objectClass.Fields()
	}
	return nil
}
func (no NilableObject) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddConstructor(NilLiteral{}, func(ValueObject) (ValueObject, error) {
		return &NilableObject{no.ClassObject, nil}, nil
	})
	innerConstructors := no.ClassObject.Constructors()
	for classHash, constructor := range innerConstructors.values {
		if constructorFn, ok := constructor.(ConstructorFn); ok {
			// FIXME: needs a better way then assigning directly to the hashmap
			csMap.values[classHash] = ConstructorFn(func(obj ValueObject) (ValueObject, error) {
				if obj == nil {
					return nil, fmt.Errorf("cannot construct %s from nil", no.ClassName())
				}
				csObj, err := constructorFn(obj)
				if err != nil {
					return nil, err
				}
				return &NilableObject{no.ClassObject, csObj}, nil
			})
		}
	}
	if genericFn := innerConstructors.generic; genericFn != nil {
		fn := *genericFn
		newFn := GenericConstructor(func(data map[string]ValueObject) (ValueObject, error) {
			csObj, err := fn(data)
			if err != nil {
				return nil, err
			}
			return &NilableObject{no.ClassObject, csObj}, nil
		})
		csMap.generic = &newFn
	}
	return csMap
}

func (no NilableObject) Class() Class {
	return no
}
func (no NilableObject) Value() interface{} {
	if no.Object == nil {
		return nil
	}
	return no.Object.Value()
}
func (no *NilableObject) Set(key string, obj ValueObject) error {
	if no.Object == nil {
		return fmt.Errorf("cannot set value on nil")
	}
	return no.Object.Set(key, obj)
}
func (no NilableObject) Get(key string) (Object, error) {
	if no.Object == nil {
		return nil, nil
	}
	return no.Object.Get(key)
}

func (no NilableObject) ComparableRules() ComparatorRules {
	rules := NewComparatorRules()
	if comparable, ok := no.ClassObject.(ComparableClass); ok {
		rules = comparable.ComparableRules()
	}
	rules.AddComparator(NilLiteral{}, tokens.EQUALS, func(a, b ValueObject) (ValueObject, error) {
		object := a.(*NilableObject)
		return BooleanLiteral(object.Object == nil), nil
	})
	rules.AddComparator(NilLiteral{}, tokens.NOT_EQUALS, func(a, b ValueObject) (ValueObject, error) {
		object := a.(*NilableObject)
		return BooleanLiteral(object.Object != nil), nil
	})
	return rules
}
func (no NilableObject) OperatorRules() OperatorRules {
	if operable, ok := no.ClassObject.(OperableClass); ok {
		return operable.OperatorRules()
	}
	return NewOperatorRules()
}

type Error struct {
	Name    string      `json:"name"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ErrorFromMapObject(obj *MapObject) Error {
	err := Error{}
	if name, ok := obj.Data["name"]; ok {
		err.Name = fmt.Sprintf("%s", name.Value())
	}
	if message, ok := obj.Data["message"]; ok {
		err.Message = fmt.Sprintf("%s", message.Value())
	}
	if data, ok := obj.Data["data"]; ok {
		err.Data = data
	}
	return err
}

func (err Error) ClassName() string {
	return err.Name
}
func (err Error) Constructors() ConstructorMap {
	return NewConstructorMap()
}

func (err Error) Class() Class {
	return err
}
func (err Error) Value() interface{} {
	return err.Message
}
func (err Error) Set(key string, obj ValueObject) error {
	return CannotSetPropertyError(key, err)
}
func (err Error) Get(key string) (Object, error) {
	return nil, nil
}

func (err Error) Error() string {
	if err.Data == nil {
		return fmt.Sprintf("%s: %s", err.Name, err.Message)
	} else {
		return fmt.Sprintf("%s: %s %s", err.Name, err.Message, err.Data)
	}
}
