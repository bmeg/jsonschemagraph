package model

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestFlattenMarshalerPreservesGeneratedRecordFormat(t *testing.T) {
	data, err := structpb.NewStruct(map[string]any{"name": "example"})
	if err != nil {
		t.Fatal(err)
	}

	marshaler := NewFlattenMarshaler()
	vertex, err := marshaler.Marshal(&Vertex{Id: "vertex-1", Label: "Patient", Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(vertex), `{"_id":"vertex-1","_label":"Patient","name":"example"}`; got != want {
		t.Fatalf("vertex JSON = %s, want %s", got, want)
	}

	edge, err := marshaler.Marshal(&Edge{Id: "edge-1", Label: "subject", From: "specimen-1", To: "patient-1", Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(edge), `{"id":"edge-1","label":"subject","from":"specimen-1","to":"patient-1","data":{"name":"example"}}`; got != want {
		t.Fatalf("edge JSON = %s, want %s", got, want)
	}
}
