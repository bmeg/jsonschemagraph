package util

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

func (s GraphSchema) Generate(classID string, data map[string]any, clean bool) ([]GraphElement, error) {
	//fmt.Println("CLASS: ", classID, "DATA: ", data)
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
				fmt.Println("VALUE OF ERROR: ", err)
				return nil, err
			}
		}
		out := make([]GraphElement, 0, 1)
		if id, nerr := getObjectID(data, class); nerr == nil {
			var ListOfRels []string
			vData := map[string]any{}
			if ext, ok := class.Extensions[GraphExtensionTag]; ok {
				gext := ext.(GraphExtension)
				// trying to index into derivedId with the appropriate json pointer patter that is taken from templatePointers
				for _, target := range gext.Targets {
					ListOfRels = append(ListOfRels, target.Rel)
					pointer_fragment := ""
					for _, pointer_string := range target.templatePointer {
						splitted_pointer := strings.Split(pointer_string.(string), "/")
						if len(splitted_pointer) < 3 {
							return nil, fmt.Errorf("length of templatePointers is not long enough")
						}
						// this if statement is used to get map[string]any into type any which is much easier to work with
						// this assumes that the first characters of the target.templatePointer will always be of the form '/hgfdsadfg/'
						if derivedId, ok := data[splitted_pointer[1]].(any); ok {
							rest_of_pointer := strings.Join(splitted_pointer[2:], "/") + "/"
							//fmt.Println("----------------------------------------------------------------", rest_of_value)
							for _, v := range rest_of_pointer {
								if v != 45 && v != 47 {
									pointer_fragment = pointer_fragment + string(v)
								} else if v == 45 {
									if _, ok := derivedId.([]any); ok {
										if value, ok := derivedId.([]any); ok {
											derivedId = value[0]
											//fmt.Println("v==45", derivedId, "VAL VALUE: ", pointer_fragment)
										}
									}
								} else if v == 47 {
									if _, ok := derivedId.(map[string]any); ok {
										if nalue, ok := derivedId.(map[string]any)[pointer_fragment]; ok {
											derivedId = nalue
											//fmt.Println("v==47", derivedId, "VAL VALUE: ", pointer_fragment)
										}
									}
									pointer_fragment = ""
								}
							}

							edgeOut := Edge{
								To:    derivedId.(string),
								From:  id,
								Label: target.Rel,
								// this label isn't quite right. Ex: right now "Transcript" -> should be "transcripts"
								// doing some string manipulation now, but not sure if there is a better solution
							}
							out = append(out, GraphElement{OutEdge: &edgeOut})
							if target.Backref != "" {
								edgeIn := Edge{
									To:    id,
									From:  derivedId.(string),
									Label: target.Backref,
								}
								out = append(out, GraphElement{InEdge: &edgeIn})
							}
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
