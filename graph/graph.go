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

	sch, err := s.Compiler.Compile(classID)
	if err != nil {
		log.Printf("compile error: %s", err)
		return nil
	}
	return sch
}
