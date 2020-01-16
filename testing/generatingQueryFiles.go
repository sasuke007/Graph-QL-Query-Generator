package testing

import (
	"log"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/ast"
)

func GetNamedType(namedType *ast.Type) string {
	if namedType.NamedType == "" {
		return namedType.Elem.NamedType
	}
	return namedType.NamedType
}

func checkCircularLoop(parent []string, field string) bool {
	for _, par := range parent {
		if par == field {
			return true
		}
	}
	return false
}

//abhi bhi circular dependency check karne ke logic me kuch dikkat h shayad
func generateFieldsForQuery(field string, complexTypes map[string][]Fields, scalarTypes []string, parent []string, enumTypes map[string]ast.EnumValueList, unionTypes map[string]*ast.Definition, interfaceTypes map[string]*ast.Definition, implementsInterface map[string]*ast.Definition) (bool, string) {
	if checkCircularLoop(parent, field) {
		return true, ""
	}
	query := ""
	parent = append(parent, field)
	typeFields, ok := complexTypes[field]
	if !ok {
		log.Fatalln("no such complex field found ", field)
	}
	for _, typeField := range typeFields {
		if _, ok := interfaceTypes[GetNamedType(typeField.Type)]; ok {
			query += "{ " + generateFieldsForInterface(GetNamedType(typeField.Type), complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface, parent) + " } "
		} else if _, ok := unionTypes[GetNamedType(typeField.Type)]; ok {
			query += "{ " + generateFieldsForUnion(GetNamedType(typeField.Type), complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface, parent) + " } "
		} else if !isScalarType(GetNamedType(typeField.Type), scalarTypes) && !isEnumType(GetNamedType(typeField.Type), enumTypes) {
			circularLoopFound, text := generateFieldsForQuery(GetNamedType(typeField.Type), complexTypes, scalarTypes, parent, enumTypes, unionTypes, interfaceTypes, implementsInterface)
			if circularLoopFound {
				continue
			} else {
				query += typeField.Name + "{ " + text + " } "
			}
		} else {
			query += typeField.Name + " "
		}
	}
	return false, query
}

func generateHeaderForQuery(argumentTypes []Fields, fieldName string, typeOfQuery string) string {
	if len(argumentTypes) == 0 {
		return strcase.ToLowerCamel(typeOfQuery) + " " + strcase.ToCamel(fieldName)
	}
	head := strcase.ToLowerCamel(typeOfQuery) + " " + strcase.ToCamel(fieldName) + "("
	for _, argumentType := range argumentTypes {
		if argumentType.Type.Elem == nil {
			head += "$" + argumentType.Name + ":" + GetNamedType(argumentType.Type) + ","
		} else {
			head += "$" + argumentType.Name + ": [" + GetNamedType(argumentType.Type) + "],"
		}
	}
	head = head[:len(head)-1] + ")"
	return head
}

func isScalarType(fieldNamedType string, scalarTypes []string) bool {
	for _, scalarType := range scalarTypes {
		if scalarType == fieldNamedType {
			return true
		}
	}
	return false
}

func generateFieldsForUnion(namedType string, complexTypes map[string][]Fields, scalarTypes []string, enumTypes map[string]ast.EnumValueList, unionTypes map[string]*ast.Definition, interfaceTypes map[string]*ast.Definition, implementsInteface map[string]*ast.Definition, parent []string) string {
	query := ""
	definiton := unionTypes[namedType]
	for _, value := range definiton.Types {
		query += "... on " + value
		temp := controller(value, complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInteface, parent)
		query += temp
	}
	return query
}

func typesImplementingThisInterface(implementsInterface map[string]*ast.Definition, name string) map[string]*ast.Definition {
	objects := make(map[string]*ast.Definition)
	for _, definition := range implementsInterface {
		for _, interfaceName := range definition.Interfaces {
			if interfaceName == name {
				objects[definition.Name] = definition
			}
		}
	}
	return objects
}

func getAdditionalFields(object *ast.Definition, interfaceTypes map[string]*ast.Definition) []Fields {
	temp := make(map[string]bool, 0)
	for _, _interface := range object.Interfaces {
		for _, field := range interfaceTypes[_interface].Fields {
			temp[field.Name] = true
		}
	}
	fields := make([]Fields, 0)
	for _, field := range object.Fields {
		if _, ok := temp[field.Name]; !ok {
			tempField := Fields{
				Name: field.Name,
				Type: field.Type,
			}
			fields = append(fields, tempField)
		}
	}
	return fields
}

func getAdditionalTypes(interfaceTypes map[string]*ast.Definition, implementsInterface map[string]*ast.Definition, name string) map[string][]Fields {
	objects := typesImplementingThisInterface(implementsInterface, name)
	temp := make(map[string][]Fields)
	for _, object := range objects {
		fields := getAdditionalFields(object, interfaceTypes)
		temp[object.Name] = fields
	}
	return temp
}

func generateFieldsForInterface(namedType string, complexTypes map[string][]Fields, scalarTypes []string, enumTypes map[string]ast.EnumValueList, unionTypes map[string]*ast.Definition, interfaceTypes map[string]*ast.Definition, implementsInterface map[string]*ast.Definition, parent []string) string {
	query := ""
	definition := interfaceTypes[namedType]
	for _, field := range definition.Fields {
		if isScalarType(GetNamedType(field.Type), scalarTypes) || isEnumType(GetNamedType(field.Type), enumTypes) {
			query += field.Name + " "
		} else {
			temp := controller(namedType, complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface, parent)
			query += " { " + temp + " }"
		}
	}

	additionalTypes := getAdditionalTypes(interfaceTypes, implementsInterface, namedType)
	for name, fields := range additionalTypes {
		tempQuery := "... on " + name + " { "
		for _, field := range fields {
			if isScalarType(GetNamedType(field.Type), scalarTypes) || isEnumType(GetNamedType(field.Type), enumTypes) {
				tempQuery += field.Name + " "
			} else {
				tempQuery += field.Name + controller(GetNamedType(field.Type), complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface, parent)
			}
		}
		tempQuery += " } "
		query += tempQuery
	}
	return query
}

func controller(namedType string, complexTypes map[string][]Fields, scalarTypes []string, enumTypes map[string]ast.EnumValueList, unionTypes map[string]*ast.Definition, interfaceTypes map[string]*ast.Definition, implementsInterface map[string]*ast.Definition, parent []string) string {
	query := ""
	if _, ok := interfaceTypes[namedType]; ok {
		query = "{ " + generateFieldsForInterface(namedType, complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface, parent) + " } "
	} else if _, ok := unionTypes[namedType]; ok {
		query = "{ " + generateFieldsForUnion(namedType, complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface, parent) + " } "
	} else {
		_, str := generateFieldsForQuery(namedType, complexTypes, scalarTypes, parent, enumTypes, unionTypes, interfaceTypes, implementsInterface)
		query = "{ " + str + " } "
	}
	return query
}

func generateQuery(complexTypes map[string][]Fields, field *ast.FieldDefinition, scalarTypes []string, typeOfQuery string, enumTypes map[string]ast.EnumValueList, unionTypes map[string]*ast.Definition, interfaceTypes map[string]*ast.Definition, implementsInteface map[string]*ast.Definition) string {
	parent := make([]string, 0)
	storeArugmentTypes := make([]Fields, 0)
	outputQuery := ""
	if field.Name == "__schema" || field.Name == "__type" {
		return ""
	}
	outputQuery = "{ "
	if field.Arguments == nil {
		outputQuery += field.Name
	} else {
		outputQuery += field.Name + "("
		for _, argument := range field.Arguments {
			storeArugmentTypes = append(storeArugmentTypes, Fields{
				Name: argument.Name,
				Type: argument.Type,
			})
			outputQuery += argument.Name + ":" + "$" + argument.Name + ","
		}
		outputQuery = outputQuery[:len(outputQuery)-1] + ")"
	}
	namedType := GetNamedType(field.Type)
	if isScalarType(namedType, scalarTypes) || isEnumType(namedType, enumTypes) {
		outputQuery += " {} "
	} else {
		query := controller(namedType, complexTypes, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInteface, parent)
		outputQuery += query
	}
	outputQuery += " }"
	head := generateHeaderForQuery(storeArugmentTypes, field.Name, typeOfQuery)
	outputQuery = head + outputQuery
	return outputQuery
}

func GenerateQueriesAndMutation(complexTypes map[string][]Fields, query *ast.Definition, scalarTypes []string, enumTypes map[string]ast.EnumValueList, unionTypes map[string]*ast.Definition, interfaceTypes map[string]*ast.Definition, implementsInterface map[string]*ast.Definition) error {
	if query != nil {
		for _, field := range query.Fields {
			query := generateQuery(complexTypes, field, scalarTypes, query.Name, enumTypes, unionTypes, interfaceTypes, implementsInterface)
			if query == "" {
				continue
			}
			err := WriteDataToFile(query, "./QueryFiles/"+field.Name+".graphql")
			if err != nil {
				return errors.Wrap(err, "unable to write data (inside queries and mutations)")
			}
		}
	}
	return nil
}

func GenerateQueryFiles(schema *ast.Schema, scalarTypes []string, complexTypes map[string][]Fields, enumTypes map[string]ast.EnumValueList, unionTypes map[string]*ast.Definition, interfaceTypes map[string]*ast.Definition, implementsInterface map[string]*ast.Definition) error {
	if err := GenerateQueriesAndMutation(complexTypes, schema.Mutation, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface); err != nil {
		return err
	}
	if err := GenerateQueriesAndMutation(complexTypes, schema.Query, scalarTypes, enumTypes, unionTypes, interfaceTypes, implementsInterface); err != nil {
		return err
	}
	return nil
}
