package util

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
)

var GraphExtensionTag = "json_schema_graph"

type GraphSchema struct {
	Classes  map[string]*jsonschema.Schema
	compiler *jsonschema.Compiler
}

func yamlLoader(s string) (io.ReadCloser, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	f := u.Path
	if runtime.GOOS == "windows" {
		f = strings.TrimPrefix(f, "/")
		f = filepath.FromSlash(f)
	}
	if strings.HasSuffix(f, ".yaml") {
		source, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Error reading file: %s", f)
			return nil, err
		}
		d := map[string]any{}
		yaml.Unmarshal(source, &d)
		schemaText, err := json.Marshal(d)
		if err != nil {
			log.Printf("Error translating file: %s", f)
			return nil, err
		}
		return io.NopCloser(strings.NewReader(string(schemaText))), nil
	}
	return os.Open(f)
}

/*
	func isEdge(s string) bool {
		if strings.Contains(s, "_definitions.yaml#/to_many") {
			return true
		} else if strings.Contains(s, "_definitions.yaml#/to_one") {
			return true
		}
		return false
	}
*/
var graphExtMeta = jsonschema.MustCompileString("graphExtMeta.json", `{"properties": {
	"anchor": {
		"type": "string",
		"format": "uri-template"
	},
	"anchorPointer": {
		"type": "string",
		"anyOf": [
			{ "format": "json-pointer" },
			{ "format": "relative-json-pointer" }
		]
	},
	"rel": {
		"anyOf": [
			{ "type": "string" },
			{
				"type": "array",
				"items": { "type": "string" },
				"minItems": 1
			}
		]
	},
	"href": {
		"type": "string",
		"format": "uri-template"
	},
	"templatePointers": {
		"type": "object",
		"additionalProperties": {
			"type": "string",
			"anyOf": [
				{ "format": "json-pointer" },
				{ "format": "relative-json-pointer" }
			]
		}
	},
	"templateRequired": {
		"type": "array",
		"items": {
			"type": "string"
		},
		"uniqueItems": true
	},
	"title": {
		"type": "string"
	},
	"description": {
		"type": "string"
	},
	"$comment": {
		"type": "string"
	}
}
}`)

type graphExtCompiler struct{}

type Target struct {
	Schema     *jsonschema.Schema
	Backref    string
	Rel        string
	Regexmatch string
	//Href  string don't currently use this
	templatePointer map[string]any
}

type GraphExtension struct {
	Targets []Target
}

func (s GraphExtension) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	//fmt.Println("graph schema validate error at ", v)
	return nil
}

func (graphExtCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if e, ok := m["links"]; ok {
		if ea, ok := e.([]any); ok {
			out := GraphExtension{Targets: []Target{}}
			for i := range ea {
				if emap, ok := ea[i].(map[string]any); ok {
					rel := ""
					if emapmap, ok := emap["rel"]; ok {
						if rel_type_check, ok := emapmap.(string); ok {
							rel = rel_type_check
						}
					}
					/*
						href := ""
						if tmp, ok := emap["href"]; ok {
							if href_type_check, ok := tmp.(string); ok {
								href = href_type_check
							}
						}
					*/

					linkKey := make(map[string]any)
					if tpmap, ok := emap["templatePointers"]; ok {
						if tp_type_check, ok := tpmap.(map[string]any); ok {
							linkKey = tp_type_check
						}
					}

					if tval, ok := emap["targetSchema"]; ok {
						if tmap, ok := tval.(map[string]any); ok {
							if ref, ok := tmap["$ref"]; ok {
								if refStr, ok := ref.(string); ok {
									backRef := ""
									regex_match := ""
									if bval, ok := emap["targetHints"]; ok {
										if hval, ok := bval.(map[string]any); ok {
											if ref, ok := hval["backref"]; ok {
												if bstr, ok := ref.(string); ok {
													backRef = bstr
												} else if bstr, ok := ref.([]any)[0].(string); ok {
													backRef = bstr
												}
											}
											if regex, ok := hval["regex_match"]; ok {
												if reg_match, ok := regex.([]any)[0].(string); ok {
													regex_match = reg_match
												}
											}
											sch, err := ctx.CompileRef(refStr, "./", false)
											if err == nil {
												out.Targets = append(out.Targets, Target{
													Schema:     sch,
													Backref:    backRef,
													Rel:        rel,
													Regexmatch: regex_match,
													//Href:            href,
													templatePointer: linkKey,
												})
											} else {
												return nil, err
											}
										}
									}
								}
							}
						}
					}
				}
			}
			return out, nil
		}
	}
	return nil, nil
}

type LoadOpt struct {
	LogError func(uri string, err error)
}

func isObjectSchema(sch *jsonschema.Schema) bool {
	if sch != nil {
		for _, i := range sch.Types {
			if i == "object" {
				return true
			}
		}
		if sch.Ref != nil {
			return isObjectSchema(sch.Ref)
		}
	}
	return false
}

func isArraySchema(sch *jsonschema.Schema) bool {
	if sch != nil {
		for _, i := range sch.Types {
			if i == "array" {
				return true
			}
		}
	}
	return false
}

func ObjectScan(sch *jsonschema.Schema) []*jsonschema.Schema {
	out := []*jsonschema.Schema{}

	isObject := isObjectSchema(sch)
	if isObject {
		out = append(out, sch)
	}

	if sch.Ref != nil {
		out = append(out, ObjectScan(sch.Ref)...)
	}

	for _, i := range sch.AnyOf {
		out = append(out, ObjectScan(i)...)
	}

	return out
}

func Load(path string, opt ...LoadOpt) (GraphSchema, error) {

	jsonschema.Loaders["file"] = yamlLoader

	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	compiler.RegisterExtension(GraphExtensionTag, graphExtMeta, graphExtCompiler{})

	info, err := os.Stat(path)
	if err != nil {
		return GraphSchema{}, err
	}

	out := GraphSchema{Classes: map[string]*jsonschema.Schema{}, compiler: compiler}
	if info.IsDir() {
		files, _ := filepath.Glob(filepath.Join(path, "*.yaml"))
		if len(files) == 0 {
			return GraphSchema{}, fmt.Errorf("no schema files found")
		}

		for _, f := range files {
			if sch, err := compiler.Compile(f); err == nil {
				if sch.Title != "" {
					out.Classes[sch.Title] = sch
				} else {
					log.Printf("Title not found: %s %#v", f, sch)
				}
			} else {
				fmt.Println("ERRORRARARAAR", err)
				for _, i := range opt {
					if i.LogError != nil {
						i.LogError(f, err)
					}
				}
			}
		}
	} else {
		if sch, err := compiler.Compile(path); err == nil {
			for _, obj := range ObjectScan(sch) {
				if obj.Title != "" {

					out.Classes[obj.Title] = obj
				}
			}
		} else {

			for _, i := range opt {
				if i.LogError != nil {
					i.LogError(path, err)
				}
			}
		}
	}
	return out, nil
}
