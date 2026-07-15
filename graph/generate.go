package graph

import (
	"bytes"
	"fmt"
	"maps"

	"github.com/bmeg/jsonschema/v6"
	"github.com/bmeg/jsonschemagraph/model"
	"github.com/bmeg/jsonschemagraph/util"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s GraphSchema) Generate(classID string, data map[string]any, extraArgs map[string]any) ([]*model.GraphElement, error) {
	namespace := extractNamespace(extraArgs)
	_, id, err := validateClassAndData(&s, classID, data, true, false)
	if err != nil {
		return nil, err
	}

	edges, mErr := s.buildEdges(classID, namespace, id, data, true)
	dataPB, err := buildVertexData(data, extraArgs)
	if err != nil {
		mErr = multierror.Append(mErr, err)
		return nil, mErr.ErrorOrNil()
	}

	out := make([]*model.GraphElement, 0, len(edges)+1)
	out = append(out, &model.GraphElement{
		Vertex: &model.Vertex{
			Id:    id,
			Label: classID,
			Data:  dataPB,
		},
	})
	for _, edge := range edges {
		out = append(out, &model.GraphElement{Edge: edge})
	}
	return out, mErr.ErrorOrNil()
}

func (s GraphSchema) GenerateEdges(classID string, data map[string]any, extraArgs map[string]any) ([]*model.Edge, error) {
	return s.GenerateEdgesWithOptions(classID, data, extraArgs, true, true)
}

func (s GraphSchema) GenerateEdgesWithOptions(classID string, data map[string]any, extraArgs map[string]any, validate bool, includeBackrefs bool) ([]*model.Edge, error) {
	return s.generateEdgesInternal(classID, data, extraArgs, validate, false, includeBackrefs)
}

func (s GraphSchema) GenerateEdgesFastWithOptions(classID string, data map[string]any, extraArgs map[string]any, validate bool, includeBackrefs bool) ([]*model.Edge, error) {
	return s.generateEdgesInternal(classID, data, extraArgs, validate, true, includeBackrefs)
}

func (s GraphSchema) BuildEdgesWithID(classID, id string, data map[string]any, extraArgs map[string]any, includeBackrefs bool) ([]*model.Edge, error) {
	namespace := extractNamespace(extraArgs)
	edges, mErr := s.buildEdges(classID, namespace, id, data, includeBackrefs)
	return edges, mErr.ErrorOrNil()
}

func (s GraphSchema) generateEdgesInternal(classID string, data map[string]any, extraArgs map[string]any, validate bool, fastValidate bool, includeBackrefs bool) ([]*model.Edge, error) {
	_, id, err := validateClassAndData(&s, classID, data, validate, fastValidate)
	if err != nil {
		return nil, err
	}
	return s.BuildEdgesWithID(classID, id, data, extraArgs, includeBackrefs)
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
func validateClassAndData(s *GraphSchema, classID string, data map[string]any, validate bool, fastValidate bool) (*jsonschema.Schema, string, error) {
	class := s.GetClass(classID)
	if class == nil {
		return nil, "", fmt.Errorf("class '%s' not found", classID)
	}
	if validate {
		var err error
		if fastValidate {
			err = class.ValidateFast(data)
		} else {
			err = class.Validate(data)
		}
		if err != nil {
			return nil, "", err
		}
	}
	id, err := util.GetObjectID(data, class)
	if err != nil {
		return nil, "", err
	}
	return class, id, nil
}

func (s GraphSchema) buildEdges(classID string, namespace uuid.UUID, id string, data map[string]any, includeBackrefs bool) ([]*model.Edge, *multierror.Error) {
	plan := (&s).GetEdgePlan(classID)
	if plan == nil {
		return nil, multierror.Append(nil, fmt.Errorf("edge plan not found for class %q", classID))
	}
	var mErr *multierror.Error
	out := make([]*model.Edge, 0, plan.EstimatedEdges)
	resolver := referenceResolver{}
	for _, rule := range plan.Rules {
		items, err := resolver.resolve(rule, data)
		if err != nil {
			mErr = multierror.Append(mErr, err)
			continue
		}
		if len(items) == 0 {
			continue
		}

		for _, ref := range items {
			refType, targetID, ok := splitReference(ref)
			if !ok {
				continue
			}
			if !rule.AllowAnyMatch && rule.MatchPrefix != refType+"/*" {
				continue
			}

			var buf bytes.Buffer
			buf.WriteString(targetID)
			buf.WriteByte('-')
			buf.WriteString(id)
			buf.WriteByte('-')
			buf.WriteString(rule.Rel)

			out = append(out, &model.Edge{
				To:    targetID,
				From:  id,
				Label: rule.Rel,
				Id:    uuid.NewSHA1(namespace, buf.Bytes()).String(),
			})

			if includeBackrefs && rule.HasBackref {
				buf.Reset()
				buf.WriteString(id)
				buf.WriteByte('-')
				buf.WriteString(targetID)
				buf.WriteByte('-')
				buf.WriteString(rule.Backref)

				out = append(out, &model.Edge{
					To:    id,
					From:  targetID,
					Label: rule.Backref,
					Id:    uuid.NewSHA1(namespace, buf.Bytes()).String(),
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

func splitReference(ref string) (string, string, bool) {
	firstSlash := -1
	for i := 0; i < len(ref); i++ {
		if ref[i] == '/' {
			firstSlash = i
			break
		}
	}
	if firstSlash <= 0 || firstSlash == len(ref)-1 {
		return "", "", false
	}
	targetType := ref[:firstSlash]
	targetID := ref[firstSlash+1:]
	for i := 0; i < len(targetID); i++ {
		if targetID[i] == '/' {
			targetID = targetID[:i]
			break
		}
	}
	if targetID == "" {
		return "", "", false
	}
	return targetType, targetID, true
}
