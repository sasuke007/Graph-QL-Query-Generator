package testing

import (
	"bytes"

	"strings"
	"text/template"

	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/ast"
)

//Details and Arguments Struct are used for Generating all the query files
type Details struct {
	QueryNameUpperCase string
	QueryNameLowerCase string
	ResponseType       string
	Query              string
	Store              *ModifiedResponseInfo
	OverWrite          []*ModifiedOverwriteInfo
	Arguments          *[]Arguments
}

type ModifiedOverwriteInfo struct {
	KeyUpperCase string
	Value        string
}

type ModifiedResponseInfo struct {
	ResponseUpperCase string
	Key               string
}

type Arguments struct {
	LowerCaseKey string
	UpperCaseKey string
	Type         string
	Value        string
}

func ConvertToUpperCase(key string) string {
	words := strings.Split(key, ".")
	temp := ""
	for _, word := range words {
		temp += strcase.ToCamel(word) + "."
	}
	return temp[:len(temp)-1]
}

func GetType(namedtype *ast.Type) string {
	if namedtype.Elem == nil {
		return GetNamedType(namedtype)
	}
	return "[]" + GetNamedType(namedtype)
}

func GetArgumentType(argumentType *ast.Type) string {
	temp := MappingScalarTypesToGoTypes(GetNamedType(argumentType))
	if argumentType.Elem == nil {
		return temp
	}
	return "[]" + temp
}

func GetValueOfWord(_type *ast.Type, word string) string {
	str := strcase.ToCamel(word)
	if _type.Elem != nil {
		str += "[0]."
	} else {
		str += "."
	}
	return str
}

func GetValueOfKey(queriesAndMutations *map[string]*ast.FieldDefinition, queryName string, dependency *OverwriteInfo, inputTypes *map[string][]Fields) (string, error) {
	words := strings.Split(dependency.Key, ".")
	str := ""
	queryInfo, ok := (*queriesAndMutations)[queryName]
	if !ok {
		return "", errors.New("no such query found " + queryName)
	}
	var dataTypeOfField string
	for _, argument := range queryInfo.Arguments {
		if argument.Name == words[0] {
			str += GetValueOfWord(argument.Type, words[0])
			dataTypeOfField = GetNamedType(argument.Type)
			break
		}
	}
	if str == "" {
		return "", errors.New("no such field found in arguments" + words[0])
	}
	for i := 1; i < len(words)-1; i++ {
		done := false
		for _, field := range (*inputTypes)[dataTypeOfField] {
			if field.Name == words[i] {
				str += GetValueOfWord(field.Type, words[i])
				dataTypeOfField = GetNamedType(field.Type)
				done = true
				break
			}
		}
		if !done {
			return "", errors.New("cannot find the field " + words[i] + " in " + dataTypeOfField)
		}
	}
	str += strcase.ToCamel(words[len(words)-1])
	return str, nil
}

func GetDataTypeOfDependency(queriesAndMutations *map[string]*ast.FieldDefinition, queryName string, dependency *OverwriteInfo, inputTypes *map[string][]Fields) (string, error) {
	words := strings.Split(dependency.Key, ".")
	queryInfo := (*queriesAndMutations)[queryName]
	var dataTypeOfField string
	temp := ""
	done := false
	temp += words[0] + "."
	for _, argument := range queryInfo.Arguments {
		if argument.Name == words[0] {
			dataTypeOfField = GetNamedType(argument.Type)
			done = true
			break
		}
	}
	if !done {
		return "", errors.New("no such key found " + temp + "in " + queryName)
	}
	check := false
	done = false
	for i := 1; i < len(words)-1; i++ {
		check = true
		temp += words[i] + "."
		for _, field := range (*inputTypes)[dataTypeOfField] {
			if field.Name == words[i] {
				dataTypeOfField = GetNamedType(field.Type)
				done = true
				break
			}
		}
	}
	if check && !done {
		return "", errors.New("no such key found " + temp + "in " + queryName)
	}
	done = false
	temp += words[len(words)-1]
	for _, field := range (*inputTypes)[dataTypeOfField] {
		if field.Name == words[len(words)-1] {
			done = true
			if field.Type.Elem == nil {
				dataTypeOfField = MappingScalarTypesToGoTypes(GetNamedType(field.Type))
			} else {
				dataTypeOfField = "[]" + MappingScalarTypesToGoTypes(GetNamedType(field.Type))
			}
		}
	}
	if !done {
		return "", errors.New("no such key found " + temp + "in " + queryName)
	}
	return dataTypeOfField, nil
}

func GetValueOfDependency(dependencyValues []interface{}, dependencyValue interface{}, dataType string) string {
	data := ""
	if dependencyValue != nil {
		if dataType == "int32" {
			data = fmt.Sprint(dependencyValue.(int))
		} else if dataType == "string" {
			temp := dependencyValue.(string)
			if temp[:1] == "$" {
				data = `(*StoreResponseFields)["` + temp[1:] + `"]`
			} else {
				data = `"` + temp + `"`
			}
		} else if dataType == "float64" {
			data = fmt.Sprint(dependencyValue.(float64))
		} else if dataType == "bool" {
			data = fmt.Sprint(dependencyValue.(bool))
		} else {
			data = "some other primitive data type found"
		}
	} else {
		if dataType == "[]int32" {
			data = "[]int32{"
			for _, value := range dependencyValues {
				data += fmt.Sprintf("%d,", value.(int))
			}
			data += "}"
		} else if dataType == "[]float64" {
			data = "[]float64{"
			for _, value := range dependencyValues {
				data += fmt.Sprintf("%f,", value.(float64))
			}
			data += "}"
		} else if dataType == "[]string" {
			data = "[]string{"
			for _, value := range dependencyValues {
				data += `"` + value.(string) + `",`
			}
			data += "}"
		} else if dataType == "[]bool" {
			data = "[]bool{"
			for _, value := range dependencyValues {
				data += fmt.Sprint(value.(bool), ",")
			}
			data += "}"
		} else {
			data = "some other data  type array"
		}
	}
	return data
}

func GenerateOperations(queriesAndMutations *map[string]*ast.FieldDefinition, queriesInfo *map[string]*QueryInfo, operations *Operations, scalarTypes []string, inputTypes *map[string][]Fields) error {
	temp, err := template.ParseFiles("./testing/templates/executingQueries.tmpl")
	if err != nil {
		return errors.Wrap(err, "unable to parse templates")
	}
	for _, operation := range operations.Operations {
		buffer := bytes.NewBuffer(make([]byte, 0))
		arguments := make([]Arguments, 0)
		for _, argument := range (*queriesAndMutations)[operation.Name].Arguments {
			temp := Arguments{
				LowerCaseKey: argument.Name,
				UpperCaseKey: strcase.ToCamel(argument.Name),
				Type:         GetArgumentType(argument.Type),
				Value:        GetFieldType(argument.Type, &scalarTypes),
			}
			arguments = append(arguments, temp)
		}
		var modifiedResponseInfo *ModifiedResponseInfo = nil
		if operation.Info != nil {
			modifiedResponseInfo = &ModifiedResponseInfo{
				ResponseUpperCase: ConvertToUpperCase(operation.Info.Response),
				Key:               operation.Info.Key,
			}
		}

		var modifiedOverwriteInfo []*ModifiedOverwriteInfo = nil
		if operation.Dependencies != nil {
			for _, dependency := range operation.Dependencies {
				key, err := GetValueOfKey(queriesAndMutations, operation.Name, dependency, inputTypes)
				if err != nil {
					return err
				}
				dataType, err := GetDataTypeOfDependency(queriesAndMutations, operation.Name, dependency, inputTypes)
				if err != nil {
					return errors.Wrap(err, "error in key")
				}
				//fmt.Println(dataType)
				temp := ModifiedOverwriteInfo{
					KeyUpperCase: key,
					Value:        GetValueOfDependency(dependency.Values, dependency.Value, dataType),
				}
				modifiedOverwriteInfo = append(modifiedOverwriteInfo, &temp)
			}
		}
		query, err := ReadFile("./QueryFiles/" + operation.Name + ".graphql")
		if err != nil {
			return errors.Wrap(err, "unable to read file "+operation.Name+".graphql")
		}
		details := Details{
			QueryNameUpperCase: strcase.ToCamel(operation.Name),
			QueryNameLowerCase: operation.Name,
			ResponseType:       GetArgumentType((*queriesAndMutations)[operation.Name].Type),
			Query:              query,
			Store:              modifiedResponseInfo,
			Arguments:          &arguments,
			OverWrite:          modifiedOverwriteInfo,
		}
		err = temp.Execute(buffer, details)
		if err != nil {
			return errors.Wrap(err, "unable to execute template"+operation.Name)
		}
		if err = WriteDataToFile(buffer.String(), "./TestingFiles/queries/"+operation.Name+".go"); err != nil {
			return errors.Wrap(err, "unable to write data to file")
		}
	}
	return nil
}
