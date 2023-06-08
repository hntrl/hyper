package symbols

import (
	"fmt"
	"reflect"
	"strings"
)

var emptyErrorType = reflect.TypeOf((*error)(nil))
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

type callback struct {
	Signature callbackSignature
	Handler   reflect.Value
}

func newCallback(cb interface{}) callback {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		panic("callback must be a function")
	}
	args := make([]reflect.Type, cbType.NumIn())
	for i := 0; i < len(args); i++ {
		args[i] = cbType.In(i)
	}
	returns := make([]reflect.Type, cbType.NumOut())
	for i := 0; i < len(returns); i++ {
		returns[i] = cbType.Out(i)
	}
	return callback{
		Signature: callbackSignature{
			args:    args,
			returns: returns,
		},
		Handler: reflect.ValueOf(cb),
	}
}

func (cb callback) AcceptsParameters(signature callbackSignature) bool {
	if len(signature.args) != len(cb.Signature.args) || len(signature.returns) != len(cb.Signature.returns) {
		return false
	}
	for idx, arg := range signature.args {
		if !cb.Signature.args[idx].ConvertibleTo(arg) {
			return false
		}
	}
	for idx, returnType := range signature.returns {
		if !cb.Signature.returns[idx].ConvertibleTo(returnType) {
			return false
		}
	}
	return true
}
func (cb callback) Call(args []reflect.Value) []reflect.Value {
	convertedArguments := make([]reflect.Value, len(args))
	for idx, arg := range args {
		convertedArguments[idx] = arg.Convert(cb.Signature.args[idx])
	}
	out := cb.Handler.Call(convertedArguments)
	convertedReturnValues := make([]reflect.Value, len(cb.Signature.returns))
	for idx, returnValue := range out {
		convertedReturnValues[idx] = returnValue.Convert(cb.Signature.returns[idx])
	}
	return convertedReturnValues
}
