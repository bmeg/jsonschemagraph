// Package model contains the graph records emitted by jsonschemagraph.
package model

import (
	"github.com/bytedance/sonic"
	"google.golang.org/protobuf/types/known/structpb"
)

// Graph is a schema graph with its vertices and edges.
type Graph struct {
	Graph    string    `json:"graph,omitempty"`
	Vertices []*Vertex `json:"vertices,omitempty"`
	Edges    []*Edge   `json:"edges,omitempty"`
}

// Vertex is a labeled graph vertex with arbitrary structured data.
type Vertex struct {
	Id    string           `json:"id,omitempty"`
	Label string           `json:"label,omitempty"`
	Data  *structpb.Struct `json:"data,omitempty"`
}

// Edge is a directed, labeled graph edge with arbitrary structured data.
type Edge struct {
	Id    string           `json:"id,omitempty"`
	Label string           `json:"label,omitempty"`
	From  string           `json:"from,omitempty"`
	To    string           `json:"to,omitempty"`
	Data  *structpb.Struct `json:"data,omitempty"`
}

// GraphElement contains one graph vertex or edge.
type GraphElement struct {
	Graph  string  `json:"graph,omitempty"`
	Vertex *Vertex `json:"vertex,omitempty"`
	Edge   *Edge   `json:"edge,omitempty"`
}

// FlattenMarshaler encodes graph records in the line-oriented format used by
// the generate command.
type FlattenMarshaler struct{}

func NewFlattenMarshaler() *FlattenMarshaler {
	return &FlattenMarshaler{}
}

func (m *FlattenMarshaler) Marshal(value any) ([]byte, error) {
	if vertex, ok := value.(*Vertex); ok {
		out := vertex.Data.AsMap()
		out["_id"] = vertex.Id
		out["_label"] = vertex.Label
		return sonic.ConfigFastest.Marshal(out)
	}
	return sonic.ConfigFastest.Marshal(value)
}
