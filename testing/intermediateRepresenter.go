package testing

import (
	"encoding/json"
	//"fmt"

	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/ast"
)

type OverwriteInfo struct {
	Key    string        `json:"key"`
	Value  interface{}   `json:"value"`
	Values []interface{} `json:"values"`
}

type ResponseInfo struct {
	Response string `json:"response"`
	Key      string `json:"key"`
}

type QueryInfo struct {
	Name         string           `json:"name"`
	Info         *ResponseInfo    `json:"info"`
	Dependencies []*OverwriteInfo `json:"dependencies"`
}

type Operations struct {
	Operations []*QueryInfo `json:"operations"`
}

func ParseIntermediateRepresenter(queriesAndMutations *map[string]*ast.FieldDefinition, inputTypes map[string][]Fields) (*Operations, map[string]*QueryInfo, error) {
	var operations Operations
	data, err := ReadFile("./intermediateRepresentation.json")
	if err != nil {
		return nil, nil, err
	}
	if err = json.Unmarshal([]byte(data), &operations); err != nil {
		return nil, nil, errors.Wrap(err, "Cannot parse intermediate Representer")
	}
	storeQueryInfo := make(map[string]*QueryInfo)
	for _, operation := range operations.Operations {
		storeQueryInfo[operation.Name] = operation
	}
	if err := validateIR(&operations, queriesAndMutations, inputTypes); err != nil {
		return nil, nil, errors.Wrap(err, "unable to validate ir")
	}
	return &operations, storeQueryInfo, nil
}

func getDataTypeOfDependencyValue(dependencyValue interface{}, dependencyValues []interface{}) (string, error) {
	dataType := ""
	var err error
	if dependencyValue != nil {
		switch dependencyValue.(type) {
		case string:
			dataType = "string"
		case float64:
			dataType = "float64"
		case bool:
			dataType = "bool"
		default:
			err = errors.New("some non primitive data type found in dependency value")
		}
	} else {
		if temp, err := getDataTypeOfDependencyValue(dependencyValues[0], nil); err != nil {
			return "", err
		} else {
			dataType = "[]" + temp
		}

	}
	return dataType, err
}

func validateIR(operations *Operations, queriesAndMutations *map[string]*ast.FieldDefinition, inputTypes map[string][]Fields) error {
	var storeKeys = make(map[string]bool)
	for _, operation := range operations.Operations {
		if _, ok := (*queriesAndMutations)[operation.Name]; !ok {
			return errors.New(operation.Name + " not a valid query or mutation")
		}
		if operation.Info != nil {
			//Info struct ki key ko verify karna hai abhi
			storeKeys[operation.Info.Key] = true
		}
		for _, dependency := range operation.Dependencies {
			dataTypeKey, err := GetDataTypeOfDependency(queriesAndMutations, operation.Name, dependency, &inputTypes)
			if err != nil {
				return err
			}
			dataTypeValue, err := getDataTypeOfDependencyValue(dependency.Value, dependency.Values)
			if err != nil {
				return err
			}
			err = errors.New("data type not matched in " + dependency.Key + " inside operation " + operation.Name)
			if dataTypeValue == "string" {
				str := dependency.Value.(string)
				if str[:1] == "$" {
					if _, ok := storeKeys[str[1:]]; !ok {
						return errors.New("unable to find key " + str + "in " + operation.Name)
					}
				}
				if dataTypeKey != dataTypeValue {
					return err
				}
			} else if dataTypeValue == "bool" {
				if dataTypeKey != dataTypeValue {
					return err
				}
			} else if dataTypeValue == "float64" {
				if !(dataTypeKey == "float64" || dataTypeKey == "int32") {
					return err
				}
			} else if dataTypeValue == "[]string" {
				if dataTypeKey != dataTypeValue {
					return err
				}
			} else if dataTypeValue == "[]bool" {
				if dataTypeKey != dataTypeValue {
					return err
				}
			} else if dataTypeValue == "[]float64" {
				if !(dataTypeKey == "[]float64" || dataTypeKey == "[]int32") {
					return err
				}
			}
		}
	}
	return nil
}
