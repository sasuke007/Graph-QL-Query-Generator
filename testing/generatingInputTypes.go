package testing

import (
	"bytes"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/ast"
	"text/template"
)

// InputTypesStruct and QeuryFieldsStruct are used to generate inputTypes.go and responseTypes.go
type InputTypesStruct struct {
	QueryName   string
	QueryFields *[]QueryFieldsStruct
}

type QueryFieldsStruct struct {
	KeyUpperCase string
	KeyLowerCase string
	Type         string
	Value        string
}

func CheckElement(namedType *ast.Type) string {
	if namedType.Elem == nil {
		return ""
	}
	return "[]"
}

func GetFieldType(fieldType *ast.Type, scalarTypes *[]string) string {
	if isScalarType(GetNamedType(fieldType), *scalarTypes) {
		if fieldType.Elem != nil {
			return "[]" + MappingScalarTypesToGoTypes(GetNamedType(fieldType)) + "{" + ScalarMappingObject[GetNamedType(fieldType)] + ",}"
		} else {
			return ScalarMappingObject[GetNamedType(fieldType)]
		}
	} else {
		//array of enums ka logic handle karna hai
		if fieldType.Elem != nil {
			return "[]" + GetNamedType(fieldType) + "{" + GetNamedType(fieldType) + "Object()}"
		} else {
			return GetNamedType(fieldType) + "Object()"
		}
	}
}

func GenerateInputTypes(types *map[string][]Fields, scalarTypes *[]string, filename string) error {
	temp, err := template.ParseFiles("./testing/templates/inputTypes.tmpl")
	if err != nil {
		return errors.Wrap(err, "error in generating input types")
	}
	buffer := bytes.NewBuffer(make([]byte, 0))

	inputTypeStruct := make([]InputTypesStruct, 0)

	for name, fields := range *types {
		queryFields := make([]QueryFieldsStruct, 0)
		for _, field := range fields {
			fieldStruct := QueryFieldsStruct{
				KeyUpperCase: strcase.ToCamel(field.Name),
				KeyLowerCase: field.Name,
				Type:         CheckElement(field.Type) + MappingScalarTypesToGoTypes(GetNamedType(field.Type)),
				Value:        GetFieldType(field.Type, scalarTypes),
			}
			queryFields = append(queryFields, fieldStruct)
		}
		inputTypeStruct = append(inputTypeStruct, InputTypesStruct{
			QueryName:   name,
			QueryFields: &queryFields,
		})
	}

	if err = temp.Execute(buffer, inputTypeStruct); err != nil {
		return errors.Wrap(err, "unable to execute template")
	}

	if err = WriteDataToFile(buffer.String(), "./TestingFiles/queries/"+filename); err != nil {
		return errors.Wrap(err, "error in writing data")
	}
	return nil
}
