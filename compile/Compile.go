package compile

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/bmeg/jsonschema"
)

type SchemaExt struct {
	//Schema  *jsonschema.Schema
	Targets []Target
}

func (s *SchemaExt) Validate(ctx *jsonschema.ValidatorContext, v any) {
	return
}

func HyperMediaVocab() *jsonschema.Vocabulary {
	url := "https://json-schema.org/draft/2020-12/links"
	schema, err := jsonschema.UnmarshalJSON(strings.NewReader(ExtMeta))
	if err != nil {
		log.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(url, schema); err != nil {
		log.Fatal(err)
	}
	sch, err := c.Compile(url)
	log.Printf("SCH HERE: %#v\n", sch)
	log.Printf("VAL OF ERR: %s", err)
	if err != nil {
		log.Fatal(err)
	}

	return &jsonschema.Vocabulary{
		URL:     url,
		Schema:  sch,
		Compile: HyperMediaCompile,
	}
}

func HyperMediaCompile(ctx *jsonschema.CompilerContext, m map[string]any) (jsonschema.SchemaExt, error) {
	links, ok := m["links"].([]any)
	if !ok {
		return nil, nil
	}
	out := SchemaExt{Targets: []Target{}}
	for _, e := range links {
		emap, ok := e.(map[string]any)
		if !ok {
			return nil, nil
		}
		jsonData, err := json.Marshal(emap)
		if err != nil {
			return nil, err
		}
		Target := Target{}
		if err := json.Unmarshal(jsonData, &Target); err != nil {
			return nil, err
		}

		compiler := jsonschema.NewCompiler()
		// Target.TargetSchema.Ref
		log.Printf("HELLO: %#v\n", Target.TargetSchema.Ref)
		if Target.TargetSchema.Ref != "" {
			sch, err := compiler.Compile(Target.TargetSchema.Ref)
			if err == nil {
				Target.Schema = sch
				out.Targets = append(out.Targets, Target)
			} else {
				log.Println("ERR: ", err)
				return nil, err
			}
		}
	}
	//log.Println("OUT: ", out)
	return &out, nil
}
