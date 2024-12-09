package schconv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bmeg/jsonschemagraph/graph"
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

func ParseIntoGraphqlSchema(relpath string, graphName string) ([]*gripql.Graph, error) {
	out, err := graph.Load(relpath)
	if err != nil {
		fmt.Errorf("Err loading schema: %s: %s\n", relpath, err)
		return nil, err
	}
	graphSchema := map[string]any{
		"vertices": []map[string]any{},
		"edges":    []map[string]any{},
		"graph":    graphName,
	}

	for _, class := range out.Classes {
		if class.Title == "CodeableReference" {
			fmt.Printf("%#v\n", class)
		}
		vertexData := make(map[string]any)
		// Since reading from schema there should be no duplicate edges
		if ext, ok := class.Extensions[compile.GraphExtensionTag]; ok {
			enumData := map[string][]string{}
			for _, target := range ext.(compile.GraphExtension).Targets {
				parts := strings.Split(target.Rel, "_")
				if len(parts) == 1 {
					vertexData[parts[0]] = target.TargetHints.RegexMatch[0][:len(target.TargetHints.RegexMatch[0])-2]
					continue
				}
				base, targetType := parts[0], parts[len(parts)-1]
				enumTitle := fmt.Sprintf("%s%s", class.Title, cases.Title(language.Und, cases.NoLower).String(base))
				vertexData[base] = enumTitle
				enumData[enumTitle] = append(enumData[enumTitle], strings.ToUpper(targetType))
			}
			if enumData != nil {
				for k, v := range enumData {
					enum := map[string]any{"data": map[string]any{k: v}, "label": "Edge", "gid": k}
					graphSchema["edges"] = append(graphSchema["edges"].([]map[string]any), enum)
				}
			}
		}

		for key, sch := range class.Properties {
			if sch.Ref != nil && sch.Ref.Title != "" && slices.Contains([]string{"Reference", "Link", "FHIRPrimitiveExtension"}, sch.Ref.Title) {
				continue
			}
			vertVal := ParseSchema(sch)
			//log.Info("FLATTENED VALUES: ", flattened_values)
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
				fmt.Println("ERR State for type: ", vertVal)
				continue
			}
		}
		vertex := map[string]any{"data": vertexData, "label": "Vertex", "gid": class.Title}
		graphSchema["vertices"] = append(graphSchema["vertices"].([]map[string]any), vertex)
	}

	// Add the Wild Card Enum that contains all classes
	graphSchema["edges"] = append(graphSchema["edges"].([]map[string]any),
		map[string]any{"data": map[string]any{"Resource": out.ListClasses()},
			"label": "Edge", "gid": "Resource"})

	expandedJSON, err := json.Marshal(graphSchema)
	if err != nil {
		fmt.Errorf("Failed to marshal expanded schema: %v", err)
	}

	//For Testing purposes
	err = os.WriteFile("graphl_vertices.json", expandedJSON, 0644)
	if err != nil {
		fmt.Errorf("Failed to write to file: %v", err)
	}
	///

	graphs := gripql.Graph{}
	json.Unmarshal(expandedJSON, &graphs)
	return []*gripql.Graph{&graphs}, nil
}
