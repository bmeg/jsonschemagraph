package graph

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmeg/jsonschema/v6"
	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bytedance/sonic"
)

func Load(path string) (*GraphSchema, error) {
	c := jsonschema.NewCompiler()
	out := &GraphSchema{
		Classes:   map[string]*jsonschema.Schema{},
		EdgePlans: map[string]*ClassEdgePlan{},
		Compiler:  c,
	}

	c.AssertFormat()
	c.RegisterFormat(&jsonschema.Format{Name: "date-time", Validate: compile.ValidateFhirDateTime})
	c.RegisterFormat(&jsonschema.Format{Name: "date", Validate: compile.ValidateFhirDate})
	c.RegisterFormat(&jsonschema.Format{Name: "binary", Validate: compile.ValidateFhirBinary})
	c.RegisterFormat(&jsonschema.Format{Name: "time", Validate: compile.ValidateFhirTime})
	c.RegisterFormat(&jsonschema.Format{Name: "uuid", Validate: compile.ValidateFhirUUID})
	c.RegisterFormat(&jsonschema.Format{Name: "uri", Validate: compile.ValidateFhirURI})

	c.AssertVocabs()
	vc, err := compile.GetHyperMediaVocab()
	if err != nil {
		return nil, err
	}

	c.RegisterVocabulary(vc)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() || filepath.Ext(path) != ".json" {
		return nil, fmt.Errorf("schema file specified is not a .json file: %s", path)
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		log.Printf("os.ReadFile(%s): %s\n", path, err)
	}
	var doc any
	err = sonic.ConfigFastest.Unmarshal(fileBytes, &doc)
	if err != nil {
		return nil, err
	}
	idstr, ok := doc.(map[string]any)["$id"].(string)
	if !ok {
		return nil, fmt.Errorf("id not in doc: %s", doc)
	}
	err = c.AddResource(idstr, doc)
	if err != nil {
		return nil, err
	}
	sch, err := c.Compile(idstr)
	if err != nil {
		return nil, err
	}
	for _, obj := range compile.ObjectScan(sch) {
		if obj.Title != "" {
			out.Classes[obj.Title] = obj
			out.EdgePlans[obj.Title] = compileEdgePlan(obj)
		}
	}
	return out, nil
}
