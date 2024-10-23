package graph

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

	"github.com/bmeg/jsonschema/v5"
	_ "github.com/bmeg/jsonschema/v5/httploader"
	"github.com/bmeg/jsonschemagraph/compile"
)

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

func YamlLoader(s string) (io.ReadCloser, error) {
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

func Load(path string, opt ...LoadOpt) (GraphSchema, error) {

	jsonschema.Loaders["file"] = YamlLoader

	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	compiler.RegisterExtension(compile.GraphExtensionTag, compile.GraphExtMeta, compile.GraphExtCompiler{})

	info, err := os.Stat(path)
	if err != nil {
		return GraphSchema{}, err
	}

	out := GraphSchema{Classes: map[string]*jsonschema.Schema{}, Compiler: compiler}
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
					log.Printf("Title not found: %s %#v\n", f, sch)
				}
			} else {
				log.Printf("compiler.Compile(%s):  %s\n", f, err)
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
			log.Printf("compiler.Compile(%s):  %s\n", path, err)
			for _, i := range opt {
				if i.LogError != nil {
					i.LogError(path, err)
				}
			}
		}
	}
	return out, nil
}
