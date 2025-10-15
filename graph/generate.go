package graph

import (
	"bytes"
	"fmt"
	"maps"
	"strings"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/jsonschema/v6"
	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bmeg/jsonschemagraph/util"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s GraphSchema) Generate(classID string, data map[string]any, extraArgs map[string]any) ([]*gripql.GraphElement, error) {
	namespace := extractNamespace(extraArgs)
	class, id, err := validateClassAndData(s, classID, data)
	if err != nil {
		return nil, err
	}

	edges, mErr := s.buildEdges(class, namespace, id, data)
	dataPB, err := buildVertexData(data, extraArgs)
	if err != nil {
		mErr = multierror.Append(mErr, err)
		return nil, mErr.ErrorOrNil()
	}

	out := make([]*gripql.GraphElement, 0, len(edges)+1)
	out = append(out, &gripql.GraphElement{
		Vertex: &gripql.Vertex{
			Id:    id,
			Label: classID,
			Data:  dataPB,
		},
	})
	out = append(out, edges...)
	return out, mErr.ErrorOrNil()
}

// extractNamespace extracts the namespace from extraArgs or uses a default, returning the UUID namespace.
func extractNamespace(extraArgs map[string]any) uuid.UUID {
	namespaceDNS, ok := extraArgs["namespace"].(string)
	if !ok {
		namespaceDNS = "calypr-public.ohsu.edu"
	}
	delete(extraArgs, "namespace")
	return uuid.NewMD5(uuid.NameSpaceDNS, []byte(namespaceDNS))
}

// validateClassAndData validates the class and data, returning the class and object ID.
func validateClassAndData(s GraphSchema, classID string, data map[string]any) (*jsonschema.Schema, string, error) {
	class := s.GetClass(classID)
	if class == nil {
		return nil, "", fmt.Errorf("class '%s' not found", classID)
	}
	if err := class.Validate(data); err != nil {
		return nil, "", err
	}
	id, err := util.GetObjectID(data, class)
	if err != nil {
		return nil, "", err
	}
	return class, id, nil
}

func (s GraphSchema) buildEdges(class *jsonschema.Schema, namespace uuid.UUID, id string, data map[string]any) ([]*gripql.GraphElement, *multierror.Error) {
	var mErr *multierror.Error
	// Preallocate with estimated capacity (1 edge + 1 backref per target)
	out := make([]*gripql.GraphElement, 0, len(class.Extensions[0].(*compile.HyperMediaExt).Targets)*2)
	for _, target := range class.Extensions[0].(*compile.HyperMediaExt).Targets {
		items, err := resolveItem(target.TemplatePointers.SplittedId, data)
		if err != nil {
			mErr = multierror.Append(mErr, err)
			continue
		}
		if len(items) == 0 {
			continue
		}

		for _, elem := range items {
			elemStr, ok := elem.(string)
			if !ok {
				mErr = multierror.Append(mErr, fmt.Errorf("expected string in resolved item, got %T", elem))
				continue
			}
			splitList := strings.Split(elemStr, "/")
			if len(splitList) < 2 {
				continue
			}

			matchPrefix := target.TargetHints.RegexMatch[0]
			if matchPrefix != (splitList[0]+"/*") && matchPrefix != "Resource/*" {
				continue
			}

			targetID := splitList[1]

			var buf bytes.Buffer
			buf.WriteString(targetID)
			buf.WriteByte('-')
			buf.WriteString(id)
			buf.WriteByte('-')
			buf.WriteString(target.Rel)

			out = append(out, &gripql.GraphElement{
				Edge: &gripql.Edge{
					To:    targetID,
					From:  id,
					Label: target.Rel,
					Id:    uuid.NewSHA1(namespace, buf.Bytes()).String(),
				},
			})

			// If backref doesn't exist, don't
			backref := target.TargetHints.Backref
			if len(backref) > 0 {
				buf.Reset()
				buf.WriteString(id)
				buf.WriteByte('-')
				buf.WriteString(targetID)
				buf.WriteByte('-')
				buf.WriteString(backref[0])

				out = append(out, &gripql.GraphElement{
					Edge: &gripql.Edge{
						To:    id,
						From:  targetID,
						Label: backref[0],
						Id:    uuid.NewSHA1(namespace, buf.Bytes()).String(),
					},
				})
			}
		}
	}
	return out, mErr
}

func buildVertexData(data, extraArgs map[string]any) (*structpb.Struct, error) {
	if extraArgs != nil {
		maps.Copy(data, extraArgs)
	}
	dataPB, err := structpb.NewStruct(data)
	if err != nil {
		return nil, err
	}
	return dataPB, nil
}

func resolveItem(pointer []string, item any) ([]any, error) {
	if len(pointer) == 0 {
		return nil, fmt.Errorf("List Pointer is of len 0")
	}
	currents := []any{item}
	for _, part := range pointer {
		var newCurrents []any
		for _, curr := range currents {
			switch c := curr.(type) {
			case map[string]any:
				if next, ok := c[part]; ok {
					newCurrents = append(newCurrents, next)
				}
			case []any:
				if part != "-" {
					return nil, fmt.Errorf("expecting '-' for list iteration in json pointer")
				}
				newCurrents = append(newCurrents, c...)
			}
		}
		if len(newCurrents) == 0 {
			return nil, nil
		}
		currents = newCurrents
	}
	return currents, nil
}
