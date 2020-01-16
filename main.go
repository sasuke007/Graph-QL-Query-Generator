package main

import (
	"github.com/sasuke007/querygeneration/testing"
	"log"
	"os"
)

func ModifyBeginning(pathToFile string, data string) error {
	// Read Write Mode
	file, err := os.OpenFile(pathToFile, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteAt([]byte(data), 0)
	if err != nil {
		return err
	}
	return nil
}

func CreateDirectories() error {
	if err := testing.CreateDirectory("./QueryFiles"); err != nil {
		return err
	}
	if err := testing.CreateDirectory("./TestingFiles"); err != nil {
		return err
	}
	if err := testing.CreateDirectory("./TestingFiles/queries"); err != nil {
		return err
	}
	return nil
}

func GenerateAdditionalFiles() error {
	if err := testing.CopyFiles("./group.graphql", "./TestingFiles/schema.graphql"); err != nil {
		return err
	}
	if err := testing.CopyFiles("./testing/storeTypes.go", "./TestingFiles/queries/storeTypes.go"); err != nil {
		return err
	}
	if err := testing.CopyFiles("./testing/generatingAst.go", "./TestingFiles/queries/generatingAst.go"); err != nil {
		return err
	}
	if err := testing.CopyFiles("./testing/fileOperations.go", "./TestingFiles/queries/fileOperations.go"); err != nil {
		return err
	}

	if err := testing.CopyFiles("./go.mod", "./TestingFiles/go.mod"); err != nil {
		return err
	}

	if err := ModifyBeginning("./TestingFiles/queries/fileOperations.go", "package queries "); err != nil {
		return err
	}

	if err := ModifyBeginning("./TestingFiles/queries/generatingAst.go", "package queries "); err != nil {
		return err
	}

	if err := ModifyBeginning("./TestingFiles/queries/storeTypes.go", "package queries "); err != nil {
		return err
	}

	if err := ModifyBeginning("./TestingFiles/go.mod", "module files                               "); err != nil {
		return err
	}

	return nil
}

func main() {
	var nameOfFile = "group.graphql"
	if err := CreateDirectories(); err != nil {
		log.Fatalln(err)
	}
	if err := GenerateAdditionalFiles(); err != nil {
		log.Fatalln(err)
	}
	fileData, err := testing.ReadFile(nameOfFile)
	if err != nil {
		log.Fatalln(err)
	}
	schema, err := testing.BuildAst(fileData)
	if err != nil {
		log.Println(err)
		return
	}
	scalarTypes := testing.StoreScalarTypes(schema)
	complexTypes := testing.StoreComplexTypes(schema)
	inputTypes := testing.StoreInputTypes(schema)
	queriesAndMutations := testing.StoreQueriesAndMutations(schema)
	enumTypes := testing.StoreEnumTypes(schema)
	unionTypes := testing.StoreUnionTypes(schema)
	interfaceTypes := testing.StoreInterfaceTypes(schema)
	implementsInterface := testing.ImplementsInterface(schema)
	if err := testing.GenerateQueryFiles(schema, scalarTypes, complexTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface); err != nil {
		log.Println("error in generating query files", err)
		return
	}
	if err := testing.GeneratingVarFiles(schema, scalarTypes, complexTypes, inputTypes, queriesAndMutations); err != nil {
		log.Println("error in generating var files", err)
		return
	}
	operations, queriesInfo, err := testing.ParseIntermediateRepresenter(&queriesAndMutations, inputTypes)
	if err != nil {
		log.Println("error in parsing intermediate representer", err)
		return
	}
	if err = testing.GenerateTestingFiles(complexTypes, scalarTypes, inputTypes, operations, queriesAndMutations, &queriesInfo, enumTypes); err != nil {
		log.Println(err)
		return
	}
	if err = testing.PostProcessor(&queriesAndMutations); err != nil {
		log.Println("unable to run post processor", err)
		return
	}
}
