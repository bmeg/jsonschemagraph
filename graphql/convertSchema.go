package graphql

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bmeg/jsonschemagraph/graph"
	"github.com/bytedance/sonic"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const JSCHEMA = "jsonSchema"
const YSCHEMA = "yamlSchema"

func ParseGraphFile(relpath string, format string, graphName string, vertexSubset []string, writeFile bool) ([]*gripql.Graph, error) {
	var graphs []*gripql.Graph
	if relpath == "" {
		return nil, fmt.Errorf("path is empty")
	}
	path, err := filepath.Abs(relpath)
	if err != nil {
		path = relpath
	}
	graphs, err = ParseIntoGraphqlSchema(path, graphName, vertexSubset, writeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph at path %s: \n%v", path, err)
	}
	return graphs, nil
}

func LowerFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

func generateQueryList(classes []string) {
	for i, v := range classes {
		classes[i] = LowerFirstLetter(classes[i]) + "(offset: Int first: Int filter: JSON sort: [SortInput]  accessibility: Accessibility = all): [" + v + "Type!]!"
	}
}

func isSlice(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func ParseIntoGraphqlSchema(relpath string, graphName string, vertexSubset []string, writeFile bool) ([]*gripql.Graph, error) {
	out, err := graph.Load(relpath)
	if err != nil {
		return nil, fmt.Errorf("Err loading schema: %s: %s\n", relpath, err)
	}
	graphSchema := map[string]any{
		"vertices": []map[string]any{},
		"graph":    graphName,
	}

	for _, class := range out.Classes {
		vertexData := make(map[string]any)
		// Need a way to skip resource because it is rendered as an edge not a vertex
		if class.Title == "Resource" {
			continue
		}
		for key, sch := range class.Properties {
			/*"Reference" is not the same as links, but it should be.
			Need to generate schema that maps links onto everything that has a reference.
			Because this behavior isn't currently the case, things like codeable reference don't get rendered because they're not currently expressed as links.*/
			if key == "links" || key == "link" || sch.Ref != nil && sch.Ref.Title != "" && slices.Contains([]string{"Reference", "FHIRPrimitiveExtension"}, sch.Ref.Title) {
				continue
			}

			value := ParseSchema(sch)

			// Fields with edges that aren't defined in our internal schema are not present in the graphql schema either
			if value == nil {
				fmt.Printf("WARNING: key %s on type %s may not be supported\n", key, class.Title)
			} else if isSlice(value) && value.([]any)[0] == "Resource" {
				value.([]any)[0] = "ResourceUnion"
				vertexData[key] = value
			} else {
				vertexData[key] = value
			}
		}

		if len(class.Extensions) > 0 {
			unionData := map[string][]string{}
			unionSeen := map[string]bool{}
			for _, target := range class.Extensions[0].(*compile.HyperMediaExt).Targets {
				parts := strings.Split(target.Rel, "_")
				RegexMatch := target.TargetHints.RegexMatch[0][:len(target.TargetHints.RegexMatch[0])-2]
				if len(parts) == 1 {
					if slices.Contains(vertexSubset, RegexMatch) {
						RegexMatch += "Type"
					} else if RegexMatch == "Resource" {
						RegexMatch += "Union"
					}
					vertexData[parts[0]] = RegexMatch
					continue
				} else if len(parts) == 2 {
					base, targetType := parts[0], parts[len(parts)-1]
					if targetType != RegexMatch {
						if value, ok := vertexData[parts[0]]; ok && value != nil {
							strValue := ""
							if slice, isSlice := value.([]any); isSlice && len(slice) > 0 {
								strValue = slice[0].(string)
							} else if s, isString := value.(string); isString {
								strValue = s
							}
							if len(strValue) >= 5 && strings.HasSuffix(strValue, "Union") {
								vertexData[parts[0]] = strValue[:len(strValue)-5]
							}
						} else {
							fmt.Printf("Key not found or value is nil: %s\n %s", parts[0], vertexData)
						}
						continue
					}
					unionTitle := fmt.Sprintf("%s%s", class.Title, cases.Title(language.Und, cases.NoLower).String(base)) + "Union"
					if _, seen := unionSeen[targetType+unionTitle]; !seen {
						vertexData[base] = unionTitle
						unionSeen[targetType+unionTitle] = true
						unionData[unionTitle] = append(unionData[unionTitle], targetType)
					}
				}
			}
			if unionData != nil {
				for k, v := range unionData {
					union := map[string]any{"data": map[string]any{k: v}, "label": "Vertex", "gid": "Union"}
					graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any), union)
				}
			}
		}

		vertex := map[string]any{"data": vertexData, "label": "Vertex", "gid": class.Title}
		if slices.Contains(vertexSubset, class.Title) {
			vertex["gid"] = class.Title + "Type"
		}
		graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any), vertex)
	}

	UnionSubset := []string{}
	for _, v := range vertexSubset {
		UnionSubset = append(UnionSubset, v)
	}

	unionResource := map[string]any{"data": map[string]any{"ResourceUnion": UnionSubset}, "label": "Vertex", "gid": "Union"}
	graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any), unionResource)

	// There needs to be a way to construct this list that is only the major nodes, preferably without hardcoding it.
	// Non obvious how to do this looking at the schema.
	generateQueryList(vertexSubset)
	graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any),
		map[string]any{"data": map[string]any{"Query": vertexSubset},
			"label": "Vertex", "gid": "Query"})

	expandedJSON, err := sonic.ConfigFastest.Marshal(graphSchema)
	if err != nil {
		fmt.Errorf("Failed to marshal expanded schema: %v", err)
	}

	if writeFile {
		err = os.WriteFile("graphl_vertices.json", expandedJSON, 0644)
		if err != nil {
			fmt.Errorf("Failed to write to file: %v", err)
		}
	}
	graphs := gripql.Graph{}
	sonic.ConfigFastest.Unmarshal(expandedJSON, &graphs)
	return []*gripql.Graph{&graphs}, nil
}
