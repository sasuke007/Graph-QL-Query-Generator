package testing

import (
	"bytes"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"text/template"
)

func GenerateResponseTypes(types *map[string][]Fields, scalarTypes *[]string, filename string) error {
	temp, err := template.ParseFiles("./testing/templates/responseTypes.tmpl")
	if err != nil {
		return errors.Wrap(err, "error in generating input types")
	}
	buffer := bytes.NewBuffer(make([]byte, 0))

	inputTypeStruct := make([]InputTypesStruct, 0)

	for name, fields := range *types {
		queryFields := make([]QueryFieldsStruct, 0)
		for _, field := range fields {
			if GetNamedType(field.Type) == "Map" {
				continue
			}
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
