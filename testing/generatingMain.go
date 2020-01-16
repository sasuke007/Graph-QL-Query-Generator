package testing

import (
	"bytes"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

func GenerateMain(operations *Operations) error {
	names := make([]string, 0)
	for _, operation := range operations.Operations {
		names = append(names, strings.ToUpper(operation.Name[:1])+operation.Name[1:])
	}
	temp, err := template.ParseFiles("./testing/templates/main.tmpl")
	if err != nil {
		return errors.Wrap(err, "unable to parse main template")
	}
	buffer := bytes.NewBuffer(make([]byte, 0))
	if err = temp.Execute(buffer, names); err != nil {
		return errors.Wrap(err, "unable to execute the main template")
	}
	if err = WriteDataToFile(buffer.String(), "./TestingFiles/main.go"); err != nil {
		return errors.Wrap(err, "unable to write data in main file")
	}
	return nil
}
