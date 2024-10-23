package compile

import (
	"encoding/json"
	"github.com/bmeg/jsonschema/v5"
	_ "github.com/bmeg/jsonschema/v5/httploader"
)

func (s GraphExtension) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	//log.Println("graph schema validate error at ", v)
	return nil
}

func (GraphExtCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	links, ok := m["links"].([]any)
	if !ok {
		return nil, nil
	}
	out := GraphExtension{Targets: []Target{}}
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
		sch, err := ctx.CompileRef(Target.TargetSchema.Ref, "./", false)
		if err == nil {
			Target.Schema = sch
			out.Targets = append(out.Targets, Target)
		}else{
			return nil, err
		}
	}
	return out, nil
}
