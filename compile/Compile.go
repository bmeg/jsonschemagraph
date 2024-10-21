package compile

import (
	"github.com/bmeg/jsonschema/v5"
	_ "github.com/bmeg/jsonschema/v5/httploader"
)

var GraphExtensionTag = "json_schema_graph"

var GraphExtMeta = jsonschema.MustCompileString("graphExtMeta.json", `{"properties": {
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

type GraphExtCompiler struct{}

type Target struct {
	Schema          *jsonschema.Schema
	Backref         string
	Rel             string
	Regexmatch      string
	TemplatePointer map[string]any
}

func (s GraphExtension) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	//fmt.Println("graph schema validate error at ", v)
	return nil
}

type GraphExtension struct {
	Targets []Target
}

func (GraphExtCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
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
													Schema:          sch,
													Backref:         backRef,
													Rel:             rel,
													Regexmatch:      regex_match,
													TemplatePointer: linkKey,
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
