package graphql

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bmeg/jsonschema/v5"
)

func jsontographlprimitiveType(returnType any) any {
	switch returnType.(type) {
	case string:
		return "String"
	case int:
		return "Int"
	case bool:
		return "Boolean"
	case float64:
		return "Float"
	case []any:
		return []any{[]any{strings.Title(returnType.([]any)[0].(string))}}
	default:
		fmt.Println("ERR State for jsontographlprimitiveType: ", returnType)
		return ""
	}
}
func ParseSchema(schema *jsonschema.Schema) any {
	/* This function traverses through the compiled json schema constructing graphql schema structures in grip form */
	vertData := make(map[string]any)

	if schema.Ref != nil &&
		schema.Ref.Title != "" {
		return schema.Ref.Title
	}

	if schema.Items2020 != nil {
		if schema.Items2020.Ref != nil &&
			schema.Items2020.Ref.Title != "" {
			if slices.Contains([]string{"Reference", "FHIRPrimitiveExtension"}, schema.Items2020.Ref.Title) {
				return nil
			}
			return []any{schema.Items2020.Ref.Title}
		}
		return ParseSchema(schema.Items2020)
	}

	if len(schema.Properties) > 0 {
		for key, property := range schema.Properties {
			if val := ParseSchema(property); val != nil {
				vertData[key] = val
			}
		}
		return vertData
	}

	// AnyOf support not implemented
	if schema.AnyOf != nil {
		return nil
	}

	return jsontographlprimitiveType(schema.Types[0])
}
