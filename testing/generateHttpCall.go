package testing

import (
	"github.com/pkg/errors"
)

func GenerateHttpCall() error {
	data, err := ReadFile("./testing/templates/httpCall.tmpl")
	if err != nil {
		return errors.Wrap(err, "unable to read httpCall.tmpl")
	}
	if err = WriteDataToFile(data, "./TestingFiles/queries/httpCall.go"); err != nil {
		return errors.Wrap(err, "unable to write data in httpCall.go")
	}
	return nil
}
