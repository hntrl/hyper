package state

import (
	"fmt"

	"github.com/hntrl/hyper/internal/symbols"
)

// Recursively flatten a given ObjectClass into a period delimited map with all fields
func flattenObject(val symbols.ValueObject, m map[string]interface{}, p string) error {
	obj, ok := val.Class().(symbols.ObjectClass)
	if !ok {
		return fmt.Errorf("cannot flatten non-object")
	}
	for k := range obj.Fields() {
		val, err := val.Get(k)
		if err != nil {
			return err
		}
		if valueObj, ok := val.(symbols.ValueObject); ok {
			if class, ok := valueObj.Class().(symbols.ObjectClass); ok && class.Fields() != nil {
				err := flattenObject(valueObj, m, p+k+".")
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
