package database

import (
	"fmt"
	"strings"
)

// pdm = period-delimited map

// FieldValues describes the values of a record in a period-delimited map
type FieldValues map[FieldName]interface{}

func NewFieldValuesFromEmbeddedMap(in map[string]interface{}) FieldValues {
	result := make(FieldValues)
	constructPDMFromMap(in, result)
	return result
}

func (f FieldValues) Map() map[string]interface{} {
	result := make(map[string]interface{})
	constructMapFromFieldValues(f, result)
	return result
}

func constructMapFromFieldValues(in FieldValues, result map[string]interface{}) {
	objectsToConstruct := make(map[string]FieldValues)
	for key, value := range in {
		keyParts := strings.Split(string(key), ".")
		if len(keyParts) != 1 {
			if _, ok := objectsToConstruct[keyParts[0]]; !ok {
				objectsToConstruct[keyParts[0]] = make(FieldValues)
			}
			objectsToConstruct[keyParts[0]][FieldName(strings.Join(keyParts[1:], "."))] = value
		} else {
			result[keyParts[0]] = value
		}
	}
	for key, value := range objectsToConstruct {
		constructedMap := make(map[string]interface{})
		constructMapFromFieldValues(value, constructedMap)
		result[key] = constructedMap
	}
}

func constructPDMFromMap(in map[string]interface{}, result FieldValues) {
	for key, value := range in {
		switch val := value.(type) {
		case map[string]interface{}:
			innerMap := make(FieldValues)
			constructPDMFromMap(val, innerMap)
			for innerKey, innerValue := range innerMap {
				result[FieldName(fmt.Sprintf("%s.%s", key, innerKey))] = innerValue
			}
		default:
			result[FieldName(key)] = value
		}
	}
}
