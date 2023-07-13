package jsgraph

import (
	"fmt"

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
		if id, nerr := getObjectID(data, class); nerr == nil {
			vData := map[string]any{}
			if ext, ok := class.Extensions[GraphExtensionTag]; ok {
				gext := ext.(GraphExtension)
				for _, target := range gext.Targets {
					if val, ok := data[target.Rel].([]any); ok {
						idTwo := ""
						// this logic needs to be edited to better reflect the schema.
						if value, ok := val[0].(any).(map[string]any); ok {
							//fmt.Println("TARGET LINK KEY ", target)
							if tmp, ok := value[target.LinkKey].(string); ok {
								idTwo = tmp
							}
						}
						edgeOut := Edge{
							To:    idTwo,
							From:  id,
							Label: target.Schema.Title, // this label isn't quite right. Ex: right now "Transcript" -> should be "transcripts"
							// could do some string manipulation but not sure if there is a better solution
						}
						out = append(out, GraphElement{OutEdge: &edgeOut})

						if target.Backref != "" {
							edgeIn := Edge{
								To:    id,
								From:  idTwo,
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
