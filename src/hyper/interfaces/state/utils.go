package state

import (
	"fmt"

	"github.com/hntrl/hyper/src/hyper/symbols"
)

// Recursively flatten a given ValueObject and serialize all properties into a period delimited map
func flattenObject(val symbols.ValueObject, m map[string]interface{}, p string) error {
	properties := val.Class().Descriptors().Properties
	if properties == nil {
		return fmt.Errorf("cannot flatten non-object")
	}
	for k, propertyAttributes := range properties {
		propertyValue, err := propertyAttributes.Getter(val)
		if err != nil {
			return err
		}
		if class := propertyValue.Class(); class.Descriptors().Properties != nil {
			err := flattenObject(propertyValue, m, p+k+".")
			if err != nil {
				return err
			}
		} else {
			if val := propertyValue.Value(); val != nil {
				m[p+k] = val
			}
		}
	}
	return nil
}
