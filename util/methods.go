package util

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
)

func (s GraphSchema) ListClasses() []string {
	out := []string{}
	for c := range s.Classes {
		out = append(out, c)
	}
	return out
}

func (s GraphSchema) GetClass(classID string) *jsonschema.Schema {
	//fmt.Println("S.CLASSES", s.Classes)
	if class, ok := s.Classes[classID]; ok {
		return class
	}
	var err error
	var sch *jsonschema.Schema
	if sch, err = s.compiler.Compile(classID); err == nil {
		return sch
	}
	//log.Printf("compile error: %s", err)
	return nil
}
