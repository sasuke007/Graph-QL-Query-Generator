package testing

import (
	"github.com/vektah/gqlparser/ast"
	//"log"
)

func MappingScalarTypesToGoTypes(choice string) string {
	var ans string
	switch choice {
	case "String":
		ans = "string"
	case "Int":
		ans = "int32"
	case "Float":
		ans = "float64"
	case "ID":
		ans = "string"
	case "Bytes":
		ans = "[]byte"
	case "Timestamp":
		ans = "string"
	case "Empty":
		ans = "interface{}"
	case "Duration":
		ans = "int32"
	case "Boolean":
		ans = "bool"
	case "Map":
		ans = "string"
	default:
		ans = choice
	}
	return ans
}

func GenerateTestingFiles(complexTypes map[string][]Fields, scalarTypes []string, inputTypes map[string][]Fields, operations *Operations, queriesAndMutations map[string]*ast.FieldDefinition, queriesInfo *map[string]*QueryInfo, enumTypes map[string]ast.EnumValueList) error {
	if err := GenerateInputTypes(&inputTypes, &scalarTypes, "inputTypes.go"); err != nil {
		return err
	}
	if err := GenerateResponseTypes(&complexTypes, &scalarTypes, "responseTypes.go"); err != nil {
		return err
	}
	if err := GenerateEnumTypes(enumTypes); err != nil {
		return err
	}
	if err := GenerateMain(operations); err != nil {
		return err
	}
	if err := GenerateHttpCall(); err != nil {
		return err
	}
	if err := GenerateOperations(&queriesAndMutations, queriesInfo, operations, scalarTypes, &inputTypes); err != nil {
		return err
	}
	return nil
}
