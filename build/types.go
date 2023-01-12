package build

import (
	"fmt"

	"github.com/hntrl/lang/language/nodes"
	"github.com/hntrl/lang/language/tokens"
)

// GenericObject represents a set of propertties without any strict bindings to
// a type
type GenericObject struct {
	fields map[string]Class       `hash:"ignore"`
	data   map[string]ValueObject `hash:"ignore"`
}

var genericHash = getHash(GenericObject{})

func NewGenericObject() *GenericObject {
	return &GenericObject{
		fields: make(map[string]Class),
		data:   make(map[string]ValueObject),
	}
}

func (obj GenericObject) ClassName() string {
	return "GenericObject"
}
func (obj GenericObject) Fields() map[string]Class {
	return obj.fields
}
func (obj GenericObject) Constructors() ConstructorMap {
	return NewConstructorMap()
}

func (obj GenericObject) Class() Class {
	return obj
}
func (obj GenericObject) Value() interface{} {
	out := make(map[string]interface{})
	for key, obj := range obj.data {
		out[key] = obj.Value()
	}
	return out
}
func (obj *GenericObject) Set(key string, new ValueObject) error {
	if class, ok := obj.fields[key]; ok {
		if !ClassEquals(class, new.Class()) {
			return fmt.Errorf("cannot assign %s to %s", new.Class().ClassName(), class.ClassName())
		}
	}
	obj.data[key] = new
	return nil
}
func (obj GenericObject) Get(key string) Object {
	return obj.data[key]
}

// Type represents a structured list of fields that are strongly controlled
type Type struct {
	Name    string
	Private bool
	Comment string
	fields  map[string]Class
}

func (t Type) ClassName() string {
	return t.Name
}
func (t Type) Fields() map[string]Class {
	return t.fields
}
func (t Type) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddGenericConstructor(t, func(data map[string]ValueObject) (ValueObject, error) {
		obj := TypeObject{t, data}
		for key, class := range t.fields {
			if nilableClass, ok := class.(NilableObject); ok {
				nilableClass.Object = data[key]
				obj.fields[key] = &nilableClass
			}
		}
		return &obj, nil
	})
	return csMap
}
func (t Type) Get(key string) Object {
	return nil
}

func (t Type) ObjectClassFromNode(ctx *Context, node nodes.ContextObject) (Class, error) {
	t.Name = node.Name
	t.Private = node.Private
	t.Comment = node.Comment

	t.fields = make(map[string]Class)

	if node.Extends != nil {
		extendsType := nodes.TypeExpression{IsArray: false, IsOptional: false, Selector: *node.Extends}
		class, err := ctx.EvaluateTypeExpression(extendsType)
		if err != nil {
			return nil, err
		}
		objectClass, ok := class.(ObjectClass)
		if !ok {
			return nil, fmt.Errorf("cannot extend %s", class.ClassName())
		}
		if fields := objectClass.Fields(); fields != nil {
			for k, v := range fields {
				t.fields[k] = v
			}
		}
	}
	for _, item := range node.Fields {
		typeExpr, ok := item.Init.(nodes.TypeStatement)
		if !ok {
			return nil, fmt.Errorf("expected type statement")
		}
		obj, err := ctx.EvaluateTypeExpression(typeExpr.Init)
		if err != nil {
			return nil, err
		}
		t.fields[typeExpr.Name] = obj
	}

	return t, nil
}

// TypeObject represents an instance of a Type
type TypeObject struct {
	ParentType Type
	fields     map[string]ValueObject
}

func (to TypeObject) Class() Class {
	return to.ParentType
}
func (to TypeObject) Value() interface{} {
	out := make(map[string]interface{})
	for key, obj := range to.fields {
		out[key] = obj.Value()
	}
	return out
}
func (to *TypeObject) Set(key string, obj ValueObject) error {
	to.fields[key] = obj
	return nil
}
func (to TypeObject) Get(key string) Object {
	return to.fields[key]
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
	// todo: implement me
	return NewConstructorMap()
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
func (it Iterable) Get(key string) Object {
	methods := map[string]Object{
		"append": NewFunction(FunctionOptions{
			Arguments: []Class{
				it.ParentType,
			},
			Returns: it,
			Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
				it.Items = append(it.Items, args[0])
				return it, nil
			},
		}),
	}
	return methods[key]
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

// TODO: Nilable types should not satisfy it's not nilable type. (i.e. String? should not satisfy String)

type NilableObject struct {
	ClassObject Class
	Object      ValueObject `hash:"ignore"`
}

func NewOptionalClass(c Class) NilableObject {
	return NilableObject{ClassObject: c}
}

func (no NilableObject) ClassName() string {
	return fmt.Sprintf("%s?", no.ClassObject.ClassName())
}
func (no NilableObject) Fields() map[string]Class {
	if objectClass, ok := no.ClassObject.(ObjectClass); ok {
		out := make(map[string]Class)
		for k, v := range objectClass.Fields() {
			out[k] = NilableObject{ClassObject: v}
		}
		return out
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
func (no NilableObject) Get(key string) Object {
	if no.Object == nil {
		return nil
	}
	return no.Object.Get(key)
}

func (no NilableObject) ComparableRules() ComparatorRules {
	rules := NewComparatorRules()
	if comparable, ok := no.ClassObject.(ComparableClass); ok {
		rules = comparable.ComparableRules()
	}
	rules.AddComparator(NilLiteral{}, tokens.EQUALS, func(a, b ValueObject) (ValueObject, error) {
		return BooleanLiteral(a == nil), nil
	})
	rules.AddComparator(NilLiteral{}, tokens.NOT_EQUALS, func(a, b ValueObject) (ValueObject, error) {
		return BooleanLiteral(a != nil), nil
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
	Name    string
	Message string
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
func (err Error) Get(key string) Object {
	return nil
}

func (err Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Name, err.Message)
}
