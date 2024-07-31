package util

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"google.golang.org/protobuf/types/known/structpb"
)

type Vertex struct {
	Gid   string           `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Label string           `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	Data  *structpb.Struct `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

type Edge struct {
	Gid   uuid.UUID        `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Label string           `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	From  string           `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	To    string           `protobuf:"bytes,4,opt,name=to,proto3" json:"to,omitempty"`
	Data  *structpb.Struct `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
}

type GraphElement struct {
	Vertex  *Vertex
	InEdge  *Edge
	OutEdge *Edge
	Field   string
}

type reference struct {
	dstID   string
	dstType string
}

func getObjectID(data map[string]any, schema *jsonschema.Schema) (string, error) {
	if id, ok := data["id"]; ok {
		if idStr, ok := id.(string); ok {
			return idStr, nil
		}
	}
	return "", fmt.Errorf("object id not found")
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func Resolve(doc interface{}, path []string) (interface{}, error) {
	var err error
	var ok bool
	curr := doc
	for _, part := range path {
		switch currTyped := curr.(type) {
		case map[string]any:
			if curr, ok = currTyped[part]; !ok {
				return nil, fmt.Errorf("key %s not found in map", part)
			}
		case []any:
			if part != "-" {
				err = fmt.Errorf("invalid list index %s", part)
			} else {
				curr = currTyped
			}
		default:
			return nil, fmt.Errorf("unable to resolve path %s on %v", part, curr)
		}
		if err != nil {
			return nil, err
		}
	}
	return curr, nil
}

func resolveItem(pointer []string, item any) ([]any, error) {
	curr := item
	var results []any
	for i, part := range pointer {
		switch currTyped := curr.(type) {
		case map[string]any:
			var err error
			var ok bool
			if curr, ok = currTyped[part]; !ok {
				return nil, fmt.Errorf("key %s not found in map", part)
			}
			// pointer miss
			if err != nil {
				return nil, err
			}
			if i == len(pointer)-1 {
				results = append(results, curr)
			}
		case []any:
			//fmt.Println("PART: ", part)
			if part == "-" {
				var tempResults []any
				for _, elem := range currTyped {
					//fmt.Println("ELEM: ", elem)
					if i == len(pointer)-1 {
						tempResults = append(tempResults, elem)
					} else {
						subPointer := pointer[i+1:]
						subResults, err := Resolve(elem, subPointer)
						//fmt.Println("ELEM: ", elem, "SUBPOINTER: ", subPointer, "SUBRESULTS: ", subResults)
						if err != nil {
							return nil, err
						}
						tempResults = append(tempResults, subResults)
					}
				}
				return tempResults, nil
			} else {
				fmt.Errorf("expecting '-' from jsonpointer")
			}
		default:
			return nil, fmt.Errorf("unable to resolve path %s on %v", part, curr)
		}
	}
	return results, nil
}

func (s GraphSchema) Generate(classID string, data map[string]any, clean bool, project_id string) ([]GraphElement, error) {
	namespace := uuid.NewMD5(uuid.NameSpaceDNS, []byte("aced-idp.org"))
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
		out := make([]GraphElement, 0, 1)
		if id, nerr := getObjectID(data, class); nerr == nil {
			var ListOfRels []string
			vData := map[string]any{}
			if ext, ok := class.Extensions[GraphExtensionTag]; ok {
				gext := ext.(GraphExtension)
				for _, target := range gext.Targets {
					ListOfRels = append(ListOfRels, target.Rel)
					pointer_string := target.templatePointer["id"]
					splitted_pointer := strings.Split(pointer_string.(string), "/")[1:]
					items, err := resolveItem(splitted_pointer, data)
					if err != nil {
						continue
					}
					for _, elem := range items {
						split_list := strings.Split(elem.(string), "/")
						if target.Regexmatch != split_list[0]+"/*" {
							continue
						}
						elem := split_list[1]
						edgeOut := Edge{
							To:    elem,
							From:  id,
							Label: target.Rel,
							Gid:   uuid.NewSHA1(namespace, []byte(fmt.Sprintf("%s-%s-%s", elem, id, target.Rel))),
						}
						out = append(out, GraphElement{OutEdge: &edgeOut})
						if target.Backref != "" {
							edgeIn := Edge{
								To:    id,
								From:  elem,
								Label: target.Backref,
								Gid:   uuid.NewSHA1(namespace, []byte(fmt.Sprintf("%s-%s-%s", id, elem, target.Backref))),
							}
							out = append(out, GraphElement{InEdge: &edgeIn})
						}
					}
				}
			}
			for name := range class.Properties {
				// gather compare to a list of rels so that the vertexes don't include edge reference information
				if !contains(ListOfRels, name) {
					if d, ok := data[name]; ok {
						vData[name] = d
					}
				}
			}
			if project_id != "" {
				project_parts := strings.Split(project_id, "-")
				if len(project_parts) != 2 {
					return nil, fmt.Errorf("project_id '%s' not in form program-project", project_id)
				}
				vData["auth_resource_path"] = "/programs/" + project_parts[0] + "/projects/" + project_parts[1]
			}

			dataPB, err := structpb.NewStruct(vData)
			if err == nil {
				vert := Vertex{Gid: id, Label: classID, Data: dataPB}
				out = append(out, GraphElement{Vertex: &vert})
			}
			if nerr != nil {
				fmt.Println("VALUE OF ERROR ", nerr) //TODO: send this to logging
			}

		}
		return out, nil
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}
