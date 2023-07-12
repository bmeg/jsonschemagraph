package jsgraph

import (
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/types/known/structpb"
)

type Vertex struct {
	Gid   string           `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
	Label string           `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	Data  *structpb.Struct `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

type Edge struct {
	Gid   string           `protobuf:"bytes,1,opt,name=gid,proto3" json:"gid,omitempty"`
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

func getReferenceIDField(data map[string]any, fieldName string) ([]reference, error) {
	out := []reference{}
	if d, ok := data[fieldName]; ok {
		//fmt.Printf("Dest id field %#v\n", d)
		if idStr, ok := d.(string); ok {
			out = append(out, reference{dstID: idStr})
		} else if idArray, ok := d.([]any); ok {
			for _, g := range idArray {
				if gStr, ok := g.(string); ok {
					out = append(out, reference{dstID: gStr})
				} else if gMap, ok := g.(map[string]any); ok {
					if id, ok := gMap["id"]; ok {
						if idStr, ok := id.(string); ok {
							out = append(out, reference{dstID: idStr})
						}
					} else if id, ok := gMap["reference"]; ok {
						//reference is a FHIR style id pointer, { "reference": "Type/id" }
						if idStr, ok := id.(string); ok {
							a := strings.Split(idStr, "/")
							if len(a) > 1 {
								out = append(out, reference{dstID: a[1], dstType: a[0]})
							}
						}
					} else {
						//fmt.Printf("id/reference Not found in %#v\n", gMap)
					}
				}
			}
		} else if idMap, ok := d.(map[string]any); ok {
			if id, ok := idMap["id"]; ok {
				if idStr, ok := id.(string); ok {
					out = append(out, reference{dstID: idStr})
				}
			} else if id, ok := idMap["reference"]; ok {
				//reference is a FHIR style id pointer, { "reference": "Type/id" }
				if idStr, ok := id.(string); ok {
					a := strings.Split(idStr, "/")
					if len(a) > 1 {
						out = append(out, reference{dstID: a[1], dstType: a[0]})
					}
				}
			}
		}
	}

	return out, nil
}

func getObjectID(data map[string]any, schema *jsonschema.Schema) (string, error) {
	if id, ok := data["id"]; ok {
		if idStr, ok := id.(string); ok {
			return idStr, nil
		}
	}
	return "", fmt.Errorf("object id not found")
}

func (s GraphSchema) Generate(classID string, data map[string]any, clean bool) ([]GraphElement, error) {
	if class := s.GetClass(classID); class != nil {
		if clean {
			var err error
			data, err = s.CleanAndValidate(class, data)
			if err != nil {
				return nil, err
			}
		} else {
			//fmt.Println("CLASS ", class)
			//fmt.Println("DATA ", data)
			err := class.Validate(data)
			if err != nil {
				return nil, err
			}
		}
		out := make([]GraphElement, 0, 1)
		// if name inside for loop inside gext object then do the work
		// iterate through links instead of properties
		// href templating lookups vs. before straight copy
		if id, nerr := getObjectID(data, class); nerr == nil {
			vData := map[string]any{}
			if ext, ok := class.Extensions[GraphExtensionTag]; ok {
				gext := ext.(GraphExtension)
				for _, target := range gext.Targets {
					if val, ok := data[target.Rel].([]any); ok {
						ToVal := ""
						if value, ok := val[0].(any).(map[string]any); ok {
							if bstr, ok := value["id"].(string); ok {
								ToVal = bstr
							}
						}
						edgeOut := Edge{
							To:    ToVal,
							From:  id,
							Label: target.Schema.Title,
						}
						out = append(out, GraphElement{OutEdge: &edgeOut})

						if target.Backref != "" {
							edgeIn := Edge{
								To:    id,
								From:  ToVal,
								Label: target.Backref,
							}
							out = append(out, GraphElement{InEdge: &edgeIn})
						}

					}
				}
			}
			for name := range class.Properties {
				//fmt.Println("NAME:        ", name)
				if d, ok := data[name]; ok {
					vData[name] = d
				}
			}

			dataPB, err := structpb.NewStruct(vData)
			if err == nil {
				vert := Vertex{Gid: id, Label: classID, Data: dataPB}
				out = append(out, GraphElement{Vertex: &vert})
			}
			if nerr != nil {
				fmt.Println("VALUE OF ERROR ", nerr)
			}

		}
		return out, nil
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}
