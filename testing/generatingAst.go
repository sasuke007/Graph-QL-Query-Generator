package testing

import (
	"github.com/pkg/errors"
	parser "github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
)

func BuildAst(inputFile string) (*ast.Schema, error) {
	myschema := ast.Source{
		Name:    "",
		Input:   inputFile,
		BuiltIn: true,
	}
	schema, err := parser.LoadSchema(&myschema)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse the file")
	}
	return schema, nil
}
