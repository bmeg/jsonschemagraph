package util

import (
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"google.golang.org/protobuf/types/known/structpb"
)

const RUNE_DASH = rune('-')
const RUNE_SLASH = rune('/')

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

// Used to filter out duplicate edges with different labels
func EdgeExistsInList(newEdge Edge, EdgeList []GraphElement) bool {
	for _, Edge := range EdgeList {
		//fmt.Println("EDGE: ", Edge, "NEWEDGE: ", newEdge)
		if Edge.OutEdge != nil && Edge.OutEdge.To != "" && Edge.OutEdge.From != "" && Edge.OutEdge.To == newEdge.To && Edge.OutEdge.From == newEdge.From {
			return true
		} else if Edge.InEdge != nil && Edge.InEdge.To != "" && Edge.InEdge.From != "" && Edge.InEdge.To == newEdge.To && Edge.InEdge.From == newEdge.From {
			return true
		}
	}
	return false
}

/*
func flattenProperties(data any, listOfRels []string, vData map[string]any) {
	switch data := data.(type) {
	case map[string]any:
		for name, value := range data {
			if !contains(listOfRels, name) {
				if d, ok := value.(map[string]any); ok {
					flattenProperties(d, listOfRels, vData)
				} else {
					if d, ok := value.(string); ok {
						fmt.Println("name: ", name, "D: ", d)
						vData[name] = d
					}
				}
			}
		}
	case []any:
		for _, element := range data {
			flattenProperties(element, listOfRels, vData)
		}
	}
}*/

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
				//fmt.Println("VALUE OF ERROR: ", err)
				return nil, err
			}
		}
		out := make([]GraphElement, 0, 1)
		if id, nerr := getObjectID(data, class); nerr == nil {
			//fmt.Println("CLASS: ", class)
			var ListOfRels []string
			vData := map[string]any{}
			//fmt.Println("HELLO ", GraphExtensionTag)
			//fmt.Println("CLASS EXTENSIONS: ", class.Extensions)
			if ext, ok := class.Extensions[GraphExtensionTag]; ok {
				gext := ext.(GraphExtension)
				// trying to index into derivedId with the appropriate json pointer patter that is taken from templatePointers
				for _, target := range gext.Targets {
					ListOfRels = append(ListOfRels, target.Rel)
					pointer_fragment := ""
					for _, pointer_string := range target.templatePointer {
						splitted_pointer := strings.Split(pointer_string.(string), "/")
						//fmt.Println("SPLITTED POINTER: ", splitted_pointer[1])

						if len(splitted_pointer) < 3 {
							return nil, fmt.Errorf("length of templatePointers is not long enough")
						}
						// this if statement is used to get map[string]any into type any which is much easier to work with
						// this assumes that the first characters of the target.templatePointer will always be of the form '/hgfdsadfg/'
						if derivedId, ok := data[splitted_pointer[1]].(any); ok {
							//fmt.Println("HELLO ", splitted_pointer, "DERIVED ID: ", derivedId)
							rest_of_pointer := strings.Join(splitted_pointer[2:], "/") + "/"
							//fmt.Println("REST OF POINTER: ", rest_of_pointer)
							//fmt.Println("----------------------------------------------------------------", rest_of_value)
							if strings.Count(rest_of_pointer, "/") > 1 {
								for _, v := range rest_of_pointer {
									if v != RUNE_DASH && v != RUNE_SLASH {
										pointer_fragment = pointer_fragment + string(v)
									} else if v == RUNE_DASH {
										if _, ok := derivedId.([]any); ok {
											if value, ok := derivedId.([]any); ok {
												if len(value) > 0 {
													derivedId = value[0]
													//fmt.Println("v==45", derivedId, "VAL VALUE: ", pointer_fragment)
												} else {
													derivedId = "" //TODO: flag validation
												}
											}
										}
									} else if v == RUNE_SLASH {
										if _, ok := derivedId.(map[string]any); ok {
											if nalue, ok := derivedId.(map[string]any)[pointer_fragment]; ok {
												derivedId = nalue
												//fmt.Println("v==47", derivedId, "VAL VALUE: ", pointer_fragment)
											}
										}
										pointer_fragment = ""
									}
								}
							} else {
								//fmt.Println("DERIVED_ID", derivedId)
								if value, ok := derivedId.(map[string]any)["id"]; ok {
									derivedId = value
								} else if value, ok := derivedId.(map[string]any)["reference"]; ok {
									derivedId = value
								}
							}
							edgeOut := Edge{
								To:    derivedId.(string),
								From:  id,
								Label: target.Rel,
								// this label isn't quite right. Ex: right now "Transcript" -> should be "transcripts"
								// doing some string manipulation now, but not sure if there is a better solution
							}
							// problem of appending edges of basically the same label giving a bunch fo duplicates?
							//if !EdgeExistsInList(edgeOut, out) {
							out = append(out, GraphElement{OutEdge: &edgeOut})
							//}
							if target.Backref != "" {
								edgeIn := Edge{
									To:    id,
									From:  derivedId.(string),
									Label: target.Backref,
								}
								//if !EdgeExistsInList(edgeIn, out) {
								out = append(out, GraphElement{InEdge: &edgeIn})
								//}
							}
						}
					}
				}
			}
			for name := range class.Properties {
				// gather compare to a list of rels so that the vertexes don't include edge reference information
				if !contains(ListOfRels, name) {
					if d, ok := data[name]; ok {
						//fmt.Println("name: ", name, "D: ", d)
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
				fmt.Println("VALUE OF ERROR ", nerr) //TODO: send this to logging
			}

		}
		return out, nil
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}
