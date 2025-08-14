package compile

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/bmeg/jsonschema/v6"
	"github.com/bytedance/sonic"
)

func GetHyperMediaVocab() (*jsonschema.Vocabulary, error) {
	c := jsonschema.NewCompiler()

	schema, err := jsonschema.UnmarshalJSON(GraphExtMeta)
	if err != nil {
		return nil, err
	}
	err = c.AddResource(GExtUrl, schema)
	if err != nil {
		return nil, err
	}
	compSch, err := c.Compile(GExtUrl)
	if err != nil {
		return nil, err
	}
	return &jsonschema.Vocabulary{
		URL:     GExtUrl,
		Schema:  compSch,
		Compile: compile,
	}, nil
}

func (s *HyperMediaExt) Validate(ctx *jsonschema.ValidatorContext, v any) {
	// This func is meant to validate the data "references" of our custom hypermedia extension.
	// I think this is already being done in the fhir data validation since we're not
	// adding any data to the vertices this can probably be left blank
}

func compile(ctx *jsonschema.CompilerContext, m map[string]any) (jsonschema.SchemaExt, error) {
	links, ok := m["links"].([]any)
	if !ok {
		return nil, nil
	}
	jsonData, err := sonic.ConfigFastest.Marshal(links)
	if err != nil {
		return nil, err
	}
	var targets []Target
	if err := sonic.ConfigFastest.Unmarshal(jsonData, &targets); err != nil {
		return nil, err
	}
	for i, target := range targets {
		if target.Rel == "" {
			return nil, &jsonschema.SchemaValidationError{
				Err: fmt.Errorf("targets[%d].rel cannot be empty", i),
			}
		}
		if target.TemplatePointers.Id != "" && !strings.HasPrefix(target.TemplatePointers.Id, "/") {
			return nil, &jsonschema.SchemaValidationError{
				Err: fmt.Errorf("targets[%d].templatePointers.id must be a valid JSON Pointer starting with '/'", i),
			}
		}
		if len(target.TargetHints.RegexMatch) > 0 {
			if target.TargetHints.RegexMatch[0] == "" {
				return nil, &jsonschema.SchemaValidationError{
					Err: fmt.Errorf("targets[%d].targetHints.regexMatch[0] cannot be an empty string", i),
				}
			}
			for j, regexStr := range target.TargetHints.RegexMatch {
				if _, err := regexp.Compile(regexStr); err != nil {
					return nil, &jsonschema.SchemaValidationError{
						Err: fmt.Errorf("targets[%d].targetHints.regexMatch[%d] is not a valid regular expression: %w", i, j, err),
					}
				}
			}
		}
	}
	return &HyperMediaExt{Targets: targets}, nil
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

func isObjectSchema(sch *jsonschema.Schema) bool {
	if sch == nil {
		return false
	}
	if sch.Types != nil && slices.Contains(sch.Types.ToStrings(), "object") {
		return true
	}
	if sch.Ref != nil {
		return isObjectSchema(sch.Ref)
	}
	return false
}
