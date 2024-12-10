package schconv

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bmeg/jsonschemagraph/graph"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ParseGraphFile(relpath string, format string, graphName string) ([]*gripql.Graph, error) {
	var graphs []*gripql.Graph
	var err error

	if relpath == "" {
		return nil, fmt.Errorf("path is empty")
	}
	// Try to get absolute path. If it fails, fall back to relative path.
	path, err := filepath.Abs(relpath)
	if err != nil {
		path = relpath
	}

	// Parse file contents
	switch format {
	case "jsonSchema":
		graphs, err = ParseIntoGraphqlSchema(path, graphName)
	case "yamlSchema":
		graphs, err = ParseIntoGraphqlSchema(relpath, graphName)
	default:
		err = fmt.Errorf("unknown file format: %s", format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph at path %s: \n%v", path, err)
	}
	return graphs, nil
}

func LowerFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	firstRune := rune(s[0])
	lowerFirst := unicode.ToLower(firstRune)
	return string(lowerFirst) + s[1:]
}

func generateQueryList(classes []string) {
	for i, v := range classes {
		classes[i] = LowerFirstLetter(classes[i]) + "(offset: Int first: Int filter: JSON sort: JSON accessibility: Accessibility = all format: Format = json): [" + v + "]"
	}
}
func ParseIntoGraphqlSchema(relpath string, graphName string) ([]*gripql.Graph, error) {
	out, err := graph.Load(relpath)
	if err != nil {
		fmt.Errorf("Err loading schema: %s: %s\n", relpath, err)
		return nil, err
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
			if sch.Ref != nil && sch.Ref.Title != "" && slices.Contains([]string{"Reference", "Link", "Link Description Object", "FHIRPrimitiveExtension"}, sch.Ref.Title) {
				continue
			}
			vertVal := ParseSchema(sch)
			switch vertVal.(type) {
			case string:
				vertexData[key] = vertVal.(string)
			case int:
				vertexData[key] = vertVal.(int)
			case bool:
				vertexData[key] = vertVal.(bool)
			case float64:
				vertexData[key] = vertVal.(float64)
			case []any:
				vertexData[key] = vertVal.([]any)
			case nil:
			default:
				log.Printf("ERR State for type: ", vertVal)
				continue
			}
		}

		if ext, ok := class.Extensions[compile.GraphExtensionTag]; ok {
			enumData := map[string][]string{}
			enumSeen := map[string]bool{}
			for _, target := range ext.(compile.GraphExtension).Targets {
				parts := strings.Split(target.Rel, "_")
				RegexMatch := target.TargetHints.RegexMatch[0][:len(target.TargetHints.RegexMatch[0])-2]
				if len(parts) == 1 {
					vertexData[parts[0]] = RegexMatch
					continue
				}
				base, targetType := parts[0], parts[len(parts)-1]
				// In places where there are in-node traversals before hitting an edge, need to
				// continue with execution to avoid creating a redundant enum.
				if targetType != RegexMatch {
					continue
				}
				enumTitle := fmt.Sprintf("%s%s", class.Title, cases.Title(language.Und, cases.NoLower).String(base)) + "Type"
				if _, seen := enumSeen[targetType+enumTitle]; !seen {
					vertexData[base] = enumTitle
					enumSeen[targetType+enumTitle] = true
					enumData[enumTitle] = append(enumData[enumTitle], strings.ToUpper(targetType))
				}
			}
			if enumData != nil {
				for k, v := range enumData {
					enum := map[string]any{"data": map[string]any{k: v}, "label": "Vertex", "gid": k}
					graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any), enum)
				}
			}
		}

		vertex := map[string]any{"data": vertexData, "label": "Vertex", "gid": class.Title}
		graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any), vertex)
	}

	// Add the Wild Card Enum that contains all classes
	enumClasses := out.ListClasses()
	for i, class := range enumClasses {
		enumClasses[i] = strings.ToUpper(class)
	}
	graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any),
		map[string]any{"data": map[string]any{"Resource": enumClasses},
			"label": "Vertex", "gid": "Resource"})

	// Add Query type so that vertices can be queried

	listClasses := out.ListClasses()
	// There needs to be a way to construct this list that is only the major nodes, preferably without hardcoding it.
	// Non obvious how to do this looking at the schema.
	generateQueryList(listClasses)
	graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any),
		map[string]any{"data": map[string]any{"Query": listClasses},
			"label": "Vertex", "gid": "Query"})

	expandedJSON, err := json.Marshal(graphSchema)
	if err != nil {
		fmt.Errorf("Failed to marshal expanded schema: %v", err)
	}

	//For Testing purposes
	/*
		err = os.WriteFile("graphl_vertices.json", expandedJSON, 0644)
		if err != nil {
			fmt.Errorf("Failed to write to file: %v", err)
			}
	*/

	graphs := gripql.Graph{}
	json.Unmarshal(expandedJSON, &graphs)
	return []*gripql.Graph{&graphs}, nil
}