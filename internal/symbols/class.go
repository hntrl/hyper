package symbols

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/hntrl/hyper/internal/tokens"

	"github.com/mitchellh/hashstructure"
)

// Converts a byte array into a ValueObject
func FromBytes(bytes []byte) (ValueObject, error) {
	var str string
	isString := json.Unmarshal(bytes, &str) == nil
	var obj interface{}
	err := json.Unmarshal(bytes, &obj)
	isJSON := err == nil

	if isJSON {
		return FromInterface(obj)
	} else if isString {
		return StringLiteral(str), nil
	} else {
		return nil, fmt.Errorf("cannot unmarshal object: %s", err)
	}
}

// Converts a standard interface into a ValueObject
func FromInterface(obj interface{}) (ValueObject, error) {
	items := reflect.ValueOf(obj)
	switch items.Kind() {
	case reflect.String:
		return StringLiteral(items.String()), nil
	case reflect.Int:
		return NumberLiteral(items.Int()), nil
	case reflect.Float64:
		return NumberLiteral(items.Float()), nil
	case reflect.Slice:
		var arr []ValueObject
		for i := 0; i < items.Len(); i++ {
			item, err := FromInterface(items.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			arr = append(arr, item)
		}
		if len(arr) == 0 {
			return Iterable{
				MapObject{},
				arr,
			}, nil
		} else {
			return Iterable{
				arr[0].Class(),
				arr,
			}, nil
		}
	case reflect.Map:
		mapObject := NewMapObject()
		iter := items.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			value, err := FromInterface(iter.Value().Interface())
			if err != nil {
				return nil, err
			}
			mapObject.Properties[key] = value.Class()
			mapObject.Data[key] = value
		}
		return mapObject, nil
	case reflect.Bool:
		return BooleanLiteral(items.Bool()), nil
	case reflect.Invalid:
		return NilLiteral{}, nil
	default:
		return nil, fmt.Errorf("unmarshal: unknown type %s", items.Kind())
	}
}

// Returns true if the two classes are equal using the hash of the classes
func ClassEquals(first, second Class) bool {
	return getHash(first) == getHash(second)
}

type classMap map[uint64]interface{}

func (m classMap) set(class Class, obj interface{}) error {
	hash, err := hashstructure.Hash(class, nil)
	if err != nil {
		return err
	}
	m[hash] = obj
	return nil
}
func (m classMap) get(class Class) interface{} {
	hash, err := hashstructure.Hash(class, nil)
	if err != nil {
		log.Fatal(err)
	}
	return m[hash]
}

func getHash(class Class) uint64 {
	hash, err := hashstructure.Hash(class, nil)
	if err != nil {
		log.Fatal(err)
	}
	return hash
}

type GenericConstructor func(map[string]ValueObject) (ValueObject, error)
type ConstructorFn func(ValueObject) (ValueObject, error)
type ConstructorMap struct {
	values  classMap
	generic *GenericConstructor
}

func NewConstructorMap() ConstructorMap {
	return ConstructorMap{values: classMap{}}
}
func (csMap *ConstructorMap) AddConstructor(class Class, fn ConstructorFn) error {
	return csMap.values.set(class, fn)
}
func (csMap *ConstructorMap) AddGenericConstructor(class ObjectClass, cb GenericConstructor) error {
	csMap.generic = &cb
	return nil
}
func (csMap ConstructorMap) Get(class Class) ConstructorFn {
	obj := csMap.values.get(class)
	if obj != nil {
		return obj.(ConstructorFn)
	}
	return nil
}
func (csMap ConstructorMap) GetGenericConstructor() *GenericConstructor {
	return csMap.generic
}

// Returns no error if a class can be constructed into another class.
func ShouldConstruct(to, from Class) error {
	if ClassEquals(to, from) {
		return nil
	}
	if _, ok := to.(AnyClass); ok {
		// If the target class is `AnyClass`, the value gets returned
		return nil
	} else if nilableClass, ok := to.(*NilableObject); ok {
		// If the target class is a nilable iterable, redo ShouldConstruct with the parent iterable as the target class
		if parentIterable, ok := nilableClass.ClassObject.(Iterable); ok {
			return ShouldConstruct(parentIterable, from)
		}
	} else if iterable, ok := to.(Iterable); ok {
		// If both the target class and the value's class are iterable, redo ShouldConstruct with their parent types
		if fromIterable, ok := from.(Iterable); ok {
			return ShouldConstruct(iterable.ParentType, fromIterable.ParentType)
		}
	}
	if nilable, ok := from.(*NilableObject); ok {
		// If the value class is optional, redo ShouldConstruct with the parent class
		return ShouldConstruct(to, nilable.ClassObject)
	} else if fn := to.Constructors().Get(from); fn != nil {
		// If the target class has a constructor defined for the value's class, it should construct
		return nil
	} else if mapFrom, ok := from.(*MapObject); ok {
		// If the value class is a MapObject, it should construct IF:
		//   1. the target class is an object class
		//   2. the target class has a generic constructor
		//   3. the map object doesn't have any properties that dont exist on the target class
		//   4. the map object isn't missing any properties that are defined on the target class (unless the property is optional)
		//   5. the map object fields are constructable to the target class's fields
		if to.Constructors().generic != nil {
			if objectClass, ok := to.(ObjectClass); ok {
				toFields := objectClass.Fields()
				// for key := range mapFrom.Properties {
				// 	if _, ok := toFields[key]; !ok {
				// 		return fmt.Errorf("unknown property %s", key)
				// 	}
				// }
				for key, class := range toFields {
					prop := mapFrom.Properties[key]
					if prop == nil {
						if _, ok := class.(*NilableObject); !ok {
							return fmt.Errorf("missing property %s", key)
						}
					} else {
						err := ShouldConstruct(class, prop)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}
		}
	} else if objectClass, ok := from.(ObjectClass); ok {
		// If the value's class is an ObjectClass and a generic constructor is defined, redo ShouldConstruct as if the value was a map object
		if fields := objectClass.Fields(); fields != nil {
			return ShouldConstruct(to, &MapObject{Properties: fields})
		}
	}
	return CannotConstructError(to.ClassName(), from.ClassName())
}

func Construct(to Class, from ValueObject) (ValueObject, error) {
	if ClassEquals(to, from.Class()) {
		return from, nil
	}
	if _, ok := to.(AnyClass); ok {
		return from, nil
	} else if nilableClass, ok := to.(*NilableObject); ok {
		if parentIterable, ok := nilableClass.ClassObject.(Iterable); ok {
			val, err := Construct(parentIterable, from)
			if err != nil {
				return nil, err
			}
			return &NilableObject{parentIterable, val}, nil
		}
	} else if iterable, ok := to.(Iterable); ok {
		if fromIterable, ok := from.(Iterable); ok {
			parentType := iterable.ParentType
			newIterable := NewIterable(parentType, len(fromIterable.Items))
			for idx, val := range fromIterable.Items {
				newItem, err := Construct(parentType, val)
				if err != nil {
					return nil, err
				}
				newIterable.Items[idx] = newItem
			}
			return newIterable, nil
		}
	}
	if nilable, ok := from.(*NilableObject); ok {
		if nilable.Object == nil {
			return nil, fmt.Errorf("cannot construct %s from nil", to.ClassName())
		}
		return Construct(to, nilable.Object)
	} else if fn := to.Constructors().Get(from.Class()); fn != nil {
		return fn(from)
	} else if mapFrom, ok := from.(*MapObject); ok {
		if genericFn := to.Constructors().generic; genericFn != nil {
			if objectClass, ok := to.(ObjectClass); ok {
				toFields := objectClass.Fields()
				// for key := range mapFrom.Properties {
				// 	if _, ok := toFields[key]; !ok {
				// 		return nil, fmt.Errorf("unknown property %s", key)
				// 	}
				// }
				fieldErrors := make(map[string]string)
				data := make(map[string]ValueObject)
				for key, class := range toFields {
					prop := mapFrom.Data[key]
					if prop == nil {
						if targetNilable, ok := class.(*NilableObject); ok {
							nilable := *targetNilable
							data[key] = &NilableObject{nilable.ClassObject, nil}
						} else {
							fieldErrors[key] = fmt.Sprintf("missing property %s", key)
						}
					} else {
						newItem, err := Construct(class, prop)
						if err != nil {
							fieldErrors[key] = err.Error()
						}
						data[key] = newItem
					}
				}
				if len(fieldErrors) > 0 {
					return nil, Error{
						Name: "ValidationError",
						Data: fieldErrors,
					}
				}
				fn := *genericFn
				return fn(data)
			}
		}
	} else if objectClass, ok := from.Class().(ObjectClass); ok {
		if fields := objectClass.Fields(); fields != nil {
			data := make(map[string]ValueObject)
			for key := range fields {
				obj, err := from.Get(key)
				if err != nil {
					return nil, err
				}
				if val, ok := obj.(ValueObject); ok {
					data[key] = val
				}
			}
			return Construct(to, &MapObject{fields, data})
		}
	}
	return nil, CannotConstructError(to.ClassName(), from.Class().ClassName())
}

type OperatorFn func(ValueObject, ValueObject) (ValueObject, error)
type OperatorMap map[tokens.Token]OperatorFn
type OperatorRules struct {
	values classMap
}

func NewOperatorRules() OperatorRules {
	return OperatorRules{values: classMap{}}
}
func (rules *OperatorRules) AddOperator(class Class, token tokens.Token, fn OperatorFn) error {
	if rules.values.get(class) == nil {
		rules.values.set(class, OperatorMap{})
	}
	newMap := rules.values.get(class)
	newMap.(OperatorMap)[token] = fn
	return rules.values.set(class, newMap)
}
func (rules OperatorRules) Get(class Class, token tokens.Token) OperatorFn {
	obj := rules.values.get(class)
	if obj != nil {
		return obj.(OperatorMap)[token]
	}
	return nil
}

type ComparatorMap map[tokens.Token]OperatorFn
type ComparatorRules struct {
	values classMap
}

func NewComparatorRules() ComparatorRules {
	return ComparatorRules{values: classMap{}}
}
func (rules *ComparatorRules) AddComparator(class Class, token tokens.Token, fn OperatorFn) error {
	if rules.values.get(class) == nil {
		rules.values.set(class, ComparatorMap{})
	}
	newMap := rules.values.get(class)
	newMap.(ComparatorMap)[token] = fn
	return rules.values.set(class, newMap)
}
func (rules ComparatorRules) Get(class Class, token tokens.Token) OperatorFn {
	obj := rules.values.get(class)
	if obj != nil {
		return obj.(ComparatorMap)[token]
	}
	return nil
}

func getOperatorFn(token tokens.Token, left, right Class) (OperatorFn, error) {
	if nilableObject, ok := right.(*NilableObject); ok {
		right = nilableObject.ClassObject
	}
	if token.IsComparableOperator() {
		if comparable, ok := left.(ComparableClass); ok {
			if fn := comparable.ComparableRules().Get(right, token); fn != nil {
				return fn, nil
			}
		}
	} else if token.IsOperator() {
		if operable, ok := left.(OperableClass); ok {
			if fn := operable.OperatorRules().Get(right, token); fn != nil {
				return fn, nil
			}
		}
	}
	return nil, fmt.Errorf("%s operator not defined between %s and %s", token, left.ClassName(), right.ClassName())
}
func ShouldOperate(token tokens.Token, left, right Class) error {
	_, err := getOperatorFn(token, left, right)
	return err
}
func Operate(token tokens.Token, left, right ValueObject) (ValueObject, error) {
	if nilableObject, ok := right.(*NilableObject); ok {
		if nilableObject.Object == nil {
			return nil, fmt.Errorf("cannot operate on nil")
		}
		right = nilableObject.Object
	}
	fn, err := getOperatorFn(token, left.Class(), right.Class())
	if err != nil {
		return nil, err
	}
	return fn(left, right)
}
