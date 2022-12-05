package build

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/hntrl/lang/language/tokens"

	"github.com/mitchellh/hashstructure"
)

/* Class helpers + definitions */

// Converts a byte array into a ValueObject
func FromBytes(bytes []byte) (ValueObject, error) {
	var str string
	isString := json.Unmarshal(bytes, &str) == nil
	var obj map[string]interface{}
	isJSON := json.Unmarshal(bytes, &obj) == nil

	if isJSON {
		return FromInterface(obj)
	} else if isString {
		return StringLiteral(str), nil
	} else {
		return nil, fmt.Errorf("cannot unmarshal %s", string(bytes))
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
		return Iterable{GenericObject{}, arr}, nil
	case reflect.Map:
		generic := NewGenericObject()
		iter := items.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			value, err := FromInterface(iter.Value().Interface())
			if err != nil {
				return nil, err
			}
			generic.fields[key] = value.Class()
			generic.data[key] = value
		}
		return generic, nil
	case reflect.Bool:
		return BooleanLiteral(items.Bool()), nil
	default:
		return nil, fmt.Errorf("unmarshal: unknown type %s", items.Kind())
	}
}

// Recursively flatten a given ObjectClass into a period delimited map with all fields
func FlattenObject(val ValueObject, m map[string]interface{}, p string) error {
	obj, ok := val.Class().(ObjectClass)
	if !ok {
		return fmt.Errorf("cannot flatten non-object")
	}

	for k := range obj.Fields() {
		val := obj.Get(k)
		if valueObj, ok := val.(ValueObject); ok {
			if class, ok := valueObj.Class().(ObjectClass); ok && class.Fields() != nil {
				err := FlattenObject(valueObj, m, p+k+".")
				if err != nil {
					return err
				}
			} else {
				if val := valueObj.Value(); val != nil {
					m[p+k] = val
				}
			}
		}
	}
	return nil
}

// Returns true if the ValueObject's class is equal to the given class
func InstanceOf(obj ValueObject, class Class) bool {
	return ClassEquals(obj.Class(), class)
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

func ShouldConstruct(class, from Class) error {
	if ClassEquals(class, from) {
		return nil
	}
	if iterable, ok := class.(Iterable); ok {
		if fromIterable, ok := from.(Iterable); ok {
			if fn := iterable.ParentType.Constructors().Get(fromIterable.ParentType); fn != nil {
				return nil
			}
		}
	} else if fn := class.Constructors().Get(from); fn != nil {
		return nil
	} else if getHash(from) == genericHash {
		if genericFn := class.Constructors().generic; genericFn != nil {
			if generic, ok := from.(GenericObject); ok {
				if objectClass, ok := class.(ObjectClass); ok {
					for key := range generic.Fields() {
						if target := objectClass.Fields()[key]; target == nil {
							return fmt.Errorf("unknown property %s", key)
						}
					}
					if _, ok := class.(NilableObject); !ok {
						for key, class := range objectClass.Fields() {
							if _, ok := class.(NilableObject); !ok {
								if target := generic.Fields()[key]; target == nil {
									return fmt.Errorf("missing property %s", key)
								}
							}
						}
					}
					return nil
				}
			}
		}
	} else if objectClass, ok := from.(ObjectClass); ok {
		if objectClass.Fields() != nil {
			if _, ok := objectClass.(GenericObject); !ok { // to avoid a cycle
				return ShouldConstruct(class, GenericObject{fields: objectClass.Fields()})
			}
		}
	}
	if nilableFrom, ok := from.(NilableObject); ok {
		return ShouldConstruct(class, nilableFrom.ClassObject)
	}
	return CannotConstructError(class.ClassName(), from.ClassName())
}
func Construct(class Class, from ValueObject) (ValueObject, error) {
	if ClassEquals(class, from.Class()) {
		return from, nil
	}
	if iterable, ok := class.(Iterable); ok {
		if fromIterable, ok := from.(Iterable); ok {
			if fn := iterable.ParentType.Constructors().Get(fromIterable.ParentType); fn != nil {
				iterable.Items = make([]ValueObject, len(fromIterable.Items))
				for idx, val := range fromIterable.Items {
					var err error
					iterable.Items[idx], err = fn(val)
					if err != nil {
						return nil, err
					}
				}
				return iterable, nil
			}
		}
	} else if fn := class.Constructors().Get(from.Class()); fn != nil {
		return fn(from)
	} else if generic, ok := from.(GenericObject); ok {
		if genericFn := class.Constructors().generic; genericFn != nil {
			if objectClass, ok := class.(ObjectClass); ok {
				var err error
				data := make(map[string]ValueObject)
				for key, val := range generic.data {
					if target := objectClass.Fields()[key]; target != nil {
						data[key], err = Construct(target, val)
						if err != nil {
							return nil, err
						}
					} else {
						return nil, fmt.Errorf("unknown property %s", key)
					}
				}
				if _, ok := class.(NilableObject); !ok {
					for key, val := range objectClass.Fields() {
						if _, ok := val.(NilableObject); !ok {
							if target := generic.data[key]; target == nil {
								return nil, fmt.Errorf("missing property %s", key)
							}
						}
					}
				}
				fn := *genericFn
				return fn(data)
			}
		}
	} else if objectClass, ok := from.Class().(ObjectClass); ok {
		if objectClass.Fields() != nil {
			if _, ok := objectClass.(GenericObject); !ok { // to avoid a cycle
				data := make(map[string]ValueObject)
				fields := objectClass.Fields()
				for key := range fields {
					if value, ok := from.Get(key).(ValueObject); ok {
						data[key] = value
					}
				}
				return Construct(class, GenericObject{fields, data})
			}
		}
	}
	if nilableFrom, ok := from.(NilableObject); ok {
		if nilableFrom.Object == nil {
			return nil, fmt.Errorf("cannot construct %s from nil", class.ClassName())
		}
		return Construct(class, nilableFrom.Object)
	}
	return nil, CannotConstructError(class.ClassName(), from.Class().ClassName())
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
	if nilableObject, ok := right.(NilableObject); ok {
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
	if nilableObject, ok := right.(NilableObject); ok {
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
