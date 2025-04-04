package graph

import (
	"fmt"
	"log"

	"github.com/bmeg/jsonschema/v5"
)

type GraphSchema struct {
	Classes  map[string]*jsonschema.Schema
	Compiler *jsonschema.Compiler
}

func (s GraphSchema) Validate(classID string, data map[string]any) error {
	class := s.GetClass(classID)
	if class != nil {
		return class.Validate(data)
	}
	return fmt.Errorf("class '%s' not found", classID)
}

func (s GraphSchema) ListClasses() []string {
	out := []string{}
	for c := range s.Classes {
		out = append(out, c)
	}
	return out
}

func (s GraphSchema) GetClass(classID string) *jsonschema.Schema {
	if class, ok := s.Classes[classID]; ok {
		return class
	}
	var err error
	var sch *jsonschema.Schema
	if sch, err = s.Compiler.Compile(classID); err == nil {
		return sch
	}
	//log.Printf("compile error: %s", err)
	return nil
}

func (s GraphSchema) CleanAndValidate(class *jsonschema.Schema, data map[string]any) (map[string]any, error) {
	if class.Ref != nil {
		return s.CleanAndValidate(class.Ref, data)
	}
	out := map[string]any{}
	for k, v := range data {
		if subCls, ok := class.Properties[k]; ok {
			if subCls == nil {
				log.Printf("Weird")
			}
			if isObjectSchema(subCls) {
				if vMap, ok := v.(map[string]any); ok {
					vn, err := s.CleanAndValidate(subCls, vMap)
					if err == nil {
						out[k] = vn
					} else {
						return nil, err
					}
				}
			} else if isArraySchema(subCls) && isObjectSchema(subCls.Items2020) {
				cls := subCls.Items2020
				if cls.Ref != nil {
					cls = cls.Ref
				}
				if vArray, ok := v.([]any); ok {
					o := []any{}
					for _, v := range vArray {
						if vMap, ok := v.(map[string]any); ok {
							l, err := s.CleanAndValidate(cls, vMap)
							if err == nil {
								o = append(o, l)
							} else {
								return nil, err
							}
						}
					}
					out[k] = o
				}
			} else {
				out[k] = v
			}
		} else {
			if class.AdditionalProperties != nil {
				if addParam, ok := class.AdditionalProperties.(bool); ok {
					if addParam {
						out[k] = v
					}
				} else if addParam, ok := class.AdditionalProperties.(*jsonschema.Schema); ok {
					if err := addParam.Validate(v); err == nil {
						out[k] = v
					}
				}
			}
		}
	}
	return out, class.Validate(out)
}
