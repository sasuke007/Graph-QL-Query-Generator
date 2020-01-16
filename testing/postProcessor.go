package testing

import (
	"github.com/vektah/gqlparser/ast"
	"os"
	"os/exec"
)

func PostProcessor(queriesAndMutations *map[string]*ast.FieldDefinition) error {
	fmtCmd := exec.Command("gofmt", "-w", ".")
	fmtCmd.Stderr = os.Stderr
	if err := fmtCmd.Run(); err != nil {
		return err
	}
	for name, _ := range *queriesAndMutations {
		cmd := exec.Command("prettier", "--parser", "graphql", "--write", "./QueryFiles/"+name+".graphql")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
