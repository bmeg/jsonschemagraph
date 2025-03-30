package graph

import (
	"fmt"
	"log"
	"strings"

	"github.com/bmeg/grip/gripql"
	_ "github.com/bmeg/jsonschema/v5/httploader"
	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bmeg/jsonschemagraph/util"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

type reference struct {
	dstID   string
	dstType string
}

func resolveItem(pointer []string, item any) ([]any, error) {
	if len(pointer) == 0 {
		return []any{item}, nil
	}
	curr := item
	part := pointer[0]
	remainingPointer := pointer[1:]

	switch currTyped := curr.(type) {
	case map[string]any:
		next, ok := currTyped[part]
		// if miss, return nil
		if !ok {
			return nil, nil
		}
		return resolveItem(remainingPointer, next)
	case []any:
		if part != "-" {
			return nil, fmt.Errorf("expecting '-' for list iteration in json pointer")
		}
		var results []any
		for _, elem := range currTyped {
			subResults, err := resolveItem(remainingPointer, elem)
			if err != nil {
				return nil, err
			}
			results = append(results, subResults...)
		}
		return results, nil
	default:
		return nil, fmt.Errorf("unable to resolve path %s on %v", part, curr)
	}
}

func (s GraphSchema) Generate(classID string, data map[string]any, clean bool, extraArgs map[string]any) ([]gripql.GraphElement, error) {
	namespaceDNS := "caliper-idp.org"
	if nms, ok := extraArgs["namespace"].(string); ok {
		namespaceDNS = nms
		delete(extraArgs, "namespace")
	}
	namespace := uuid.NewMD5(uuid.NameSpaceDNS, []byte(namespaceDNS))
	if class := s.GetClass(classID); class != nil {
		if clean {
			var err error
			data, err = s.CleanAndValidate(class, data)
			if err != nil {
				return nil, err
			}
		} else {
			err := class.Validate(data)
			if err != nil {
				return nil, err
			}
		}
		out := make([]gripql.GraphElement, 0, 1)
		if id, nerr := util.GetObjectID(data, class); nerr == nil {
			vData := map[string]any{}
			if ext, ok := class.Extensions[compile.GraphExtensionTag]; ok {
				gext := ext.(compile.GraphExtension)
				for _, target := range gext.Targets {
					if target.TemplatePointers.Id == "" {
						continue
					}
					splitted_pointer := strings.Split(target.TemplatePointers.Id, "/")[1:]
					items, err := resolveItem(splitted_pointer, data)
					// if pointer miss continue
					if items == nil && err == nil {
						continue
					}
					// if invalid pointer structure in data, error
					if err != nil {
						log.Fatal("Resolve item Error: ", err)
					}
					for _, elem := range items {
						split_list := strings.Split(elem.(string), "/")
						regex_match := target.TargetHints.RegexMatch[0]
						if target.TargetHints.RegexMatch != nil && (regex_match == (split_list[0]+"/*") || regex_match == "Resource/*") {
							elem := split_list[1]
							edgeOut := gripql.Edge{
								To:    elem,
								From:  id,
								Label: target.Rel,
								Id:    uuid.NewSHA1(namespace, []byte(fmt.Sprintf("%s-%s-%s", elem, id, target.Rel))).String(),
							}
							out = append(out, gripql.GraphElement{Edge: &edgeOut})
							if target.TargetHints.Backref[0] != "" {
								edgeIn := gripql.Edge{
									To:    id,
									From:  elem,
									Label: target.TargetHints.Backref[0],
									Id:    uuid.NewSHA1(namespace, []byte(fmt.Sprintf("%s-%s-%s", id, elem, target.TargetHints.Backref[0]))).String(),
								}
								out = append(out, gripql.GraphElement{Edge: &edgeIn})
							}
						}
					}
				}
			}
			for name := range class.Properties {
				if d, ok := data[name]; ok {
					vData[name] = d
				}
			}
			if extraArgs != nil {
				for key, val := range extraArgs {
					vData[key] = val
				}
			}
			dataPB, err := structpb.NewStruct(vData)
			if err != nil {
				log.Printf("Error when creating structpb with data: %#v: %s\n", err)
				return nil, err
			}
			vert := gripql.Vertex{Id: id, Label: classID, Data: dataPB}
			out = append(out, gripql.GraphElement{Vertex: &vert})

		} else if nerr != nil {
			log.Println("Error: ", nerr)
			return nil, nerr
		}
		return out, nil
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}
