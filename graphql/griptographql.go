package graphql

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/bmeg/grip/gripql"
)

func IsUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func GripGraphqltoGraphql(graph *gripql.Graph) string {
	var schemaBuilder strings.Builder
	// Write gen3 style boiler plate to mirror thier args
	schemaBuilder.WriteString("scalar JSON\n")
	schemaBuilder.WriteString("enum Accessibility {\n  all\n  accessible\n  unaccessible\n}\n")
	schemaBuilder.WriteString("input SortInput {\n  field: String!\n  descending: Boolean\n}\n")

	for _, v := range graph.Vertices {
		//fmt.Println("V: ", v)
		//fmt.Printf("BREAK: \n\n")
		if v.Gid != "Query" {
			executedFirstBlock := false
			for name, values := range v.Data.AsMap() {
				listVals, ok := values.([]any)
				if ok && v.Gid == "Union" {
					executedFirstBlock = true
					schemaBuilder.WriteString(fmt.Sprintf("union %s =", name))
					listValslen := len(listVals)
					for i, value := range listVals {
						if i == (listValslen - 1) {
							schemaBuilder.WriteString(fmt.Sprintf(" %sType", value))
						} else {
							schemaBuilder.WriteString(fmt.Sprintf(" %sType |", value))
						}
					}
					schemaBuilder.WriteString("\n")
				} else {
					break
				}
			}
			if !executedFirstBlock {
				schemaBuilder.WriteString(fmt.Sprintf("type %s {\n", v.Gid))
				for field, fieldType := range v.Data.AsMap() {
					strFieldType, ok := fieldType.(string)
					if ok && (strings.HasSuffix(strFieldType, "Type") || strings.HasSuffix(strFieldType, "Union")) {
						schemaBuilder.WriteString(fmt.Sprintf("  %s(offset: Int first: Int): %s!\n", field, strFieldType))
					} else {
						schemaBuilder.WriteString(fmt.Sprintf("  %s: %s\n", field, fieldType))
					}
				}
				schemaBuilder.WriteString("  auth_resource_path: String\n")
				schemaBuilder.WriteString("}\n")
			}
		} else {
			for name, values := range v.Data.AsMap() {
				schemaBuilder.WriteString(fmt.Sprintf("type %s {\n", name))
				for _, value := range values.([]any) {
					schemaBuilder.WriteString(fmt.Sprintf("  %s\n", value))
				}
				schemaBuilder.WriteString("}\n")
			}
		}
	}

	fileName := "schema.graphql"
	err := os.WriteFile(fileName, []byte(schemaBuilder.String()), 0644)
	if err != nil {
		fmt.Printf("Failed to write schema to file: %v\n", err)
	}

	return schemaBuilder.String()
}
