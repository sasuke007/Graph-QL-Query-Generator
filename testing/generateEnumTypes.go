package testing

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/ast"
	"strings"
	"text/template"
)

type EnumValuesStruct struct {
	Name          string
	NameUpperCase string
	Fields        []string
	FirstField    string
}

func GenerateEnumTypes(enumTypes map[string]ast.EnumValueList) error {
	temp, err := template.ParseFiles("./testing/templates/enum.tmpl")
	if err != nil {
		return errors.Wrap(err, "unable to parse templates")
	}
	buffer := bytes.NewBuffer(make([]byte, 0))
	enumValueStruct := make([]EnumValuesStruct, 0)

	for name, fields := range enumTypes {
		if name == "__DirectiveLocation" || name == "__TypeKind" {
			continue
		}
		str := make([]string, 0)
		for _, field := range fields {

			str = append(str, field.Name)

		}
		enum := EnumValuesStruct{
			Name:          name,
			Fields:        str,
			FirstField:    str[0],
			NameUpperCase: strings.ToUpper(name),
		}
		enumValueStruct = append(enumValueStruct, enum)
	}
	if err := temp.Execute(buffer, enumValueStruct); err != nil {
		return errors.Wrap(err, "unable to execute template")
	}
	if err = WriteDataToFile(buffer.String(), "./TestingFiles/queries/enumTypes.go"); err != nil {
		errors.Wrap(err, "unable to write data to file")
	}
	return nil
}
