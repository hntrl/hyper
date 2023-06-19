package symbols

import (
	"fmt"
	"reflect"
	"strings"
)

var emptyErrorType = reflect.TypeOf((*error)(nil)).Elem()
var emptyBoolType = reflect.TypeOf(false)
var emptyIntType = reflect.TypeOf(int(0))

type callbackSignature struct {
	args    []reflect.Type
	returns []reflect.Type
}

func (cs callbackSignature) String() string {
	argNames := make([]string, len(cs.args))
	for idx, arg := range cs.args {
		argNames[idx] = arg.Name()
	}
	returnNames := make([]string, len(cs.returns))
	for idx, returnType := range cs.returns {
		returnNames[idx] = returnType.Name()
	}
	return fmt.Sprintf("func(%s) (%s)", strings.Join(argNames, ", "), strings.Join(returnNames, ", "))
}
func (cs callbackSignature) Check(cb interface{}) error {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		return fmt.Errorf("callback must be a function")
	}
	if cbType.NumIn() != len(cs.args) {
		return fmt.Errorf("callback must accept %d arguments", len(cs.args))
	}
	if cbType.NumOut() != len(cs.returns) {
		return fmt.Errorf("callback must return %d values", len(cs.returns))
	}
	for idx, arg := range cs.args {
		if !checkReflectValueEquality(arg, cbType.In(idx)) {
			return fmt.Errorf("callback argument %d must be %s", idx, arg.Name())
		}
	}
	for idx, returnType := range cs.returns {
		if !checkReflectValueEquality(returnType, cbType.Out(idx)) {
			return fmt.Errorf("callback return value %d must be %s", idx, returnType.Name())
		}
	}
	return nil
}

func checkReflectValueEquality(a, b reflect.Type) bool {
	if a.Kind() == reflect.Interface && b.Kind() == reflect.Interface {
		if !b.AssignableTo(a) {
			return false
		}
	} else if a.Kind() == reflect.Interface {
		if !b.Implements(a) {
			return false
		}
	} else {
		if !b.AssignableTo(a) {
			return false
		}
	}
	return true
}
