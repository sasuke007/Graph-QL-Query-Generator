package testing

import (
	"github.com/vektah/gqlparser/ast"
)

type Fields struct {
	Name string
	Type *ast.Type
}

type EnumFields struct {
	Name   string
	Values ast.EnumValueList
}

func StoreScalarTypes(schema *ast.Schema) []string {
	scalarTypes := make([]string, 0)
	for _, field := range schema.Types {
		if field.Kind == ast.DefinitionKind("SCALAR") {
			scalarTypes = append(scalarTypes, field.Name)
		}
	}
	return scalarTypes
}

func checkTypes(schema *ast.Schema) []*ast.Definition {
	var types []*ast.Definition
	astTypes := []string{
		"String",
		"Boolean",
		"__Type",
		"Float",
		"__Field",
		"__EnumValue",
		"Query",
		"Int",
		"ID",
		"__Schema",
		"__DirectiveLocation",
		"__InputValue",
		"__TypeKind",
		"__Directive",
		"Mutation",
	}
	for _, definition := range schema.Types {
		want := true
		for _, value := range astTypes {
			if definition.Name == value {
				want = false
				break
			}
		}
		if want {
			types = append(types, definition)
		}
	}
	return types
}

func StoreComplexTypes(schema *ast.Schema) map[string][]Fields {
	storeTypes := make(map[string][]Fields)
	types := checkTypes(schema)
	for _, definition := range types {
		if definition.Kind == ast.DefinitionKind("OBJECT") {
			storeFields := make([]Fields, 0)
			fields := definition.Fields
			for _, field := range fields {
				storeFields = append(storeFields, Fields{
					Name: field.Name,
					Type: field.Type,
				})
			}
			storeTypes[definition.Name] = storeFields
		}
	}
	return storeTypes
}

func StoreInputTypes(schema *ast.Schema) map[string][]Fields {
	storeTypes := make(map[string][]Fields)
	definitions := checkTypes(schema)
	for _, definition := range definitions {
		if definition.Kind == ast.DefinitionKind("INPUT_OBJECT") {
			storeFields := make([]Fields, 0)
			for _, field := range definition.Fields {
				storeFields = append(storeFields, Fields{
					Name: field.Name,
					Type: field.Type,
				})
			}
			storeTypes[definition.Name] = storeFields
		}
	}
	return storeTypes
}

func StoreQueriesAndMutations(schema *ast.Schema) map[string]*ast.FieldDefinition {
	storeQueriesANdMutations := make(map[string]*ast.FieldDefinition)
	if schema.Query != nil {
		for _, field := range schema.Query.Fields {
			if field.Name != "__type" && field.Name != "__schema" {
				storeQueriesANdMutations[field.Name] = field
			}
		}
	}
	if schema.Mutation != nil {
		for _, field := range schema.Mutation.Fields {
			if field.Name != "__type" && field.Name != "__schema" {
				storeQueriesANdMutations[field.Name] = field
			}
		}
	}
	return storeQueriesANdMutations
}

func StoreEnumTypes(schema *ast.Schema) map[string]ast.EnumValueList {
	enums := make(map[string]ast.EnumValueList, 0)
	for name, types := range schema.Types {
		if types.Kind == ast.DefinitionKind("ENUM") {
			enums[name] = types.EnumValues
		}
	}
	return enums
}

func isEnumType(namedType string, enumTypes map[string]ast.EnumValueList) bool {
	_, ok := enumTypes[namedType]
	return ok
}

func StoreUnionTypes(schema *ast.Schema) map[string]*ast.Definition {
	storeUnions := make(map[string]*ast.Definition)
	for name, definition := range schema.Types {
		if definition.Kind == ast.DefinitionKind("UNION") {
			storeUnions[name] = definition
		}
	}
	return storeUnions
}

func StoreInterfaceTypes(schema *ast.Schema) map[string]*ast.Definition {
	storeInterface := make(map[string]*ast.Definition)
	for name, definition := range schema.Types {
		if definition.Kind == ast.DefinitionKind("INTERFACE") {
			storeInterface[name] = definition
		}
	}
	return storeInterface
}

func ImplementsInterface(schema *ast.Schema) map[string]*ast.Definition {
	implementsInterface := make(map[string]*ast.Definition)
	for name, definition := range schema.Types {
		if definition.Interfaces != nil {
			implementsInterface[name] = definition
		}
	}
	return implementsInterface
}
