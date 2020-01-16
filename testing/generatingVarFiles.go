package testing

import (
	"github.com/pkg/errors"
	"github.com/tidwall/pretty"
	"github.com/vektah/gqlparser/ast"
)

var ScalarMappingObject = map[string]string{
	"String":    `"Demo String"`,
	"Int":       "1",
	"ID":        `""`,
	"Float":     "2.005",
	"Boolean":   "true",
	"Timestamp": `"2019-05-26T00:00:00.000Z"`,
	"Empty":     `"empty not handled"`,
	"Map":       `"e30="`,
	"Duration":  "45",
	"Bytes":     `[]byte("010100110101001000101")`,
}

func isNamedType(namedType *ast.Type) bool {
	return namedType.Elem == nil
}

func generateFields(inputTypes map[string][]Fields, scalarTypes []string, namedType string, checkNamedType bool) string {
	data := ""
	if checkNamedType {
		if isScalarType(namedType, scalarTypes) {
			data += ScalarMappingObject[namedType]
		} else {
			data += "{"
			for _, field := range inputTypes[namedType] {
				data += "\"" + field.Name + "\":"
				data += generateFields(inputTypes, scalarTypes, GetNamedType(field.Type), isNamedType(field.Type))
				data += ","
			}
			data = data[:len(data)-1]
			data += "}"
		}
	} else {
		if isScalarType(namedType, scalarTypes) {
			data += "[" + ScalarMappingObject[namedType] + "]"
		} else {
			data += "{"
			for _, field := range inputTypes[namedType] {
				data += "\"" + field.Name + "\":"
				data += "[" + generateFields(inputTypes, scalarTypes, GetNamedType(field.Type), isNamedType(field.Type)) + "],"
				data += ","
			}
			data = data[:len(data)-1]
			data += "}"
		}
	}
	return data
}

//Assumeing their is no circular dependency in input types
func generateArguments(inputTypes map[string][]Fields, scalarTypes []string, arguments ast.ArgumentDefinitionList) string {
	data := ""
	if arguments == nil {
		return data
	}
	for _, argument := range arguments {
		data += "\"" + argument.Name + "\":"
		data += generateFields(inputTypes, scalarTypes, GetNamedType(argument.Type), isNamedType(argument.Type))
		data += ","
	}
	return data[:len(data)-1]
}

func GeneratingVarFiles(schema *ast.Schema, scalarTypes []string, complexTypes map[string][]Fields, inputTypes map[string][]Fields, queriesAndMutations map[string]*ast.FieldDefinition) error {
	for query, definition := range queriesAndMutations {
		fileData := "{" + generateArguments(inputTypes, scalarTypes, definition.Arguments) + "}"

		fileData = string(pretty.Pretty([]byte(fileData)))

		if err := WriteDataToFile(fileData, "./QueryFiles/"+query+".var.json"); err != nil {
			return errors.Wrap(err, "unable to write data to file (.var.json file)")
		}
	}
	return nil
}
