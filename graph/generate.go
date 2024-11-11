package graph

import (
	"fmt"
	"log"
	"strings"

	"github.com/bmeg/jsonschemagraph/util"
	"github.com/google/uuid"
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

func (s GraphSchema) Generate(classID string, data map[string]any, clean bool, project_id string) ([]GraphElement, error) {
	//namespace := uuid.NewMD5(uuid.NameSpaceDNS, []byte("aced-idp.org"))
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
		if id, nerr := util.GetObjectID(data, class); nerr == nil {
			//var ListOfRels []string
			vData := map[string]any{}
			/*for _, target := range class.Targets {
			ListOfRels = append(ListOfRels, target.Rel)
			if target.TemplatePointers.Id == "" {
				continue
			}
			//log.Println(" TARGET TEMPLATE POINTER ID: ", target.TemplatePointers.Id )
			splitted_pointer := strings.Split(target.TemplatePointers.Id, "/")[1:]
			items, err := resolveItem(splitted_pointer, data)
			// if pointer miss continue
			if items == nil && err == nil {
				continue
			}
			// if invalid pointer structure in data, error
			if err != nil {
				log.Fatal("ERROR: ", err)
			}
			for _, elem := range items {
				split_list := strings.Split(elem.(string), "/")
				if target.TargetHints.RegexMatch != nil && target.TargetHints.RegexMatch[0] == (split_list[0]+"/*") {
					elem := split_list[1]
					edgeOut := Edge{
						To:    elem,
						From:  id,
						Label: target.Rel,
						Gid:   uuid.NewSHA1(namespace, []byte(fmt.Sprintf("%s-%s-%s", elem, id, target.Rel))),
					}
					out = append(out, GraphElement{OutEdge: &edgeOut})
					if target.TargetHints.Backref[0] != "" {
						edgeIn := Edge{
							To:    id,
							From:  elem,
							Label: target.TargetHints.Backref[0],
							Gid:   uuid.NewSHA1(namespace, []byte(fmt.Sprintf("%s-%s-%s", id, elem, target.TargetHints.Backref[0]))),
						}
						out = append(out, GraphElement{InEdge: &edgeIn})
					}
				}
			}

			}*/
			for name := range class.Properties {
				if d, ok := data[name]; ok {
					vData[name] = d
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
			if err != nil {
				log.Println("ERROR: ", err)
				return nil, err
			}
			vert := Vertex{Gid: id, Label: classID, Data: dataPB}
			out = append(out, GraphElement{Vertex: &vert})

		} else if nerr != nil {
			log.Println("ERROR: ", nerr)
			return nil, nerr
		}
		return out, nil
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}
