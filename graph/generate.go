package graph

import (
	"fmt"
	"maps"
	"strings"

	"github.com/bmeg/grip/gripql"
	_ "github.com/bmeg/jsonschema/v5/httploader"
	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bmeg/jsonschemagraph/util"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/protobuf/types/known/structpb"
)

type reference struct {
	dstID   string
	dstType string
}

/*
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
	}*/

func resolveItem(pointer []string, item any) ([]any, error) {
	if len(pointer) == 0 {
		return []any{item}, nil
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

func (s GraphSchema) Generate(classID string, data map[string]any, extraArgs map[string]any) ([]*gripql.GraphElement, error) {
	namespaceDNS, ok := extraArgs["namespace"].(string)
	if !ok {
		return nil, fmt.Errorf("Expecting Namespace for UUID seed")
	}
	delete(extraArgs, "namespace")
	namespace := uuid.NewMD5(uuid.NameSpaceDNS, []byte(namespaceDNS))

	class := s.GetClass(classID)
	if class == nil {
		return nil, fmt.Errorf("class '%s' not found", classID)
	}

	err := class.Validate(data)
	if err != nil {
		return nil, err
	}

	out := make([]*gripql.GraphElement, 0, 1)
	id, err := util.GetObjectID(data, class)
	if err != nil {
		return nil, err
	}
	ge, ok := class.Extensions[compile.GraphExtensionTag].(compile.GraphExtension)
	if !ok {
		return nil, fmt.Errorf("Expecting %s to be indexable and of type (compile.GraphExtension)s", compile.GraphExtensionTag)
	}

	var mErr *multierror.Error
	for _, target := range ge.Targets {
		if target.TemplatePointers.Id == "" {
			continue
		}
		splitted_pointer := strings.Split(target.TemplatePointers.Id, "/")[1:]
		items, err := resolveItem(splitted_pointer, data)
		if err != nil {
			mErr = multierror.Append(mErr, err)
			continue
		}

		for _, elem := range items {
			split_list := strings.Split(elem.(string), "/")
			regex_match := target.TargetHints.RegexMatch[0]
			if target.TargetHints.RegexMatch != nil && (regex_match == (split_list[0]+"/*") || regex_match == "Resource/*") {
				elem := split_list[1]
				out = append(out, &gripql.GraphElement{
					Edge: &gripql.Edge{
						To:    elem,
						From:  id,
						Label: target.Rel,
						Id:    uuid.NewSHA1(namespace, fmt.Appendf(nil, "%s-%s-%s", elem, id, target.Rel)).String(),
					}})
				if target.TargetHints.Backref[0] != "" {
					out = append(out, &gripql.GraphElement{
						Edge: &gripql.Edge{
							To:    id,
							From:  elem,
							Label: target.TargetHints.Backref[0],
							Id:    uuid.NewSHA1(namespace, fmt.Appendf(nil, "%s-%s-%s", id, elem, target.TargetHints.Backref[0])).String(),
						}})
				}
			}
		}
	}

	vData := make(map[string]any, len(class.Properties)+len(extraArgs))
	for name := range class.Properties {
		if d, ok := data[name]; ok {
			vData[name] = d
		}
	}
	if extraArgs != nil {
		maps.Copy(vData, extraArgs)
	}
	dataPB, err := structpb.NewStruct(vData)
	if err != nil {
		mErr = multierror.Append(mErr, err)
		return nil, mErr.ErrorOrNil()
	}
	out = append(out,
		&gripql.GraphElement{
			Vertex: &gripql.Vertex{
				Id:    id,
				Label: classID,
				Data:  dataPB,
			},
		},
	)
	return out, mErr.ErrorOrNil()
}
