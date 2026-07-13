package graph

import "fmt"

type referenceResolver struct {
	current []any
	next    []any
	out     []string
}

func (r *referenceResolver) resolve(rule EdgeRule, data map[string]any) ([]string, error) {
	switch rule.PointerKey {
	case "/subject/reference":
		return r.singleNested(data, "subject", "reference"), nil
	case "/specimen/reference":
		return r.singleNested(data, "specimen", "reference"), nil
	case "/patient/reference":
		return r.singleNested(data, "patient", "reference"), nil
	case "/study/reference":
		return r.singleNested(data, "study", "reference"), nil
	case "/focus/-/reference":
		return r.listNested(data, "focus", "reference"), nil
	case "/member/-/entity/reference":
		return r.memberEntityReferences(data), nil
	default:
		return r.resolveGeneric(rule.Path, data)
	}
}

func (r *referenceResolver) singleNested(data map[string]any, parent, child string) []string {
	r.out = r.out[:0]
	parentValue, ok := data[parent].(map[string]any)
	if !ok {
		return nil
	}
	value, ok := parentValue[child].(string)
	if !ok || value == "" {
		return nil
	}
	r.out = append(r.out, value)
	return r.out
}

func (r *referenceResolver) listNested(data map[string]any, field, child string) []string {
	r.out = r.out[:0]
	items, ok := data[field].([]any)
	if !ok {
		return nil
	}
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		value, ok := entry[child].(string)
		if !ok || value == "" {
			continue
		}
		r.out = append(r.out, value)
	}
	if len(r.out) == 0 {
		return nil
	}
	return r.out
}

func (r *referenceResolver) memberEntityReferences(data map[string]any) []string {
	r.out = r.out[:0]
	items, ok := data["member"].([]any)
	if !ok {
		return nil
	}
	for _, item := range items {
		member, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entity, ok := member["entity"].(map[string]any)
		if !ok {
			continue
		}
		value, ok := entity["reference"].(string)
		if !ok || value == "" {
			continue
		}
		r.out = append(r.out, value)
	}
	if len(r.out) == 0 {
		return nil
	}
	return r.out
}

func (r *referenceResolver) resolveGeneric(path []string, item any) ([]string, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("list pointer is of len 0")
	}
	r.current = r.current[:0]
	r.next = r.next[:0]
	r.out = r.out[:0]
	r.current = append(r.current, item)
	for _, part := range path {
		r.next = r.next[:0]
		for _, current := range r.current {
			switch value := current.(type) {
			case map[string]any:
				if next, ok := value[part]; ok {
					r.next = append(r.next, next)
				}
			case []any:
				if part != "-" {
					return nil, fmt.Errorf("expecting '-' for list iteration in json pointer")
				}
				r.next = append(r.next, value...)
			}
		}
		if len(r.next) == 0 {
			r.current = r.current[:0]
			return nil, nil
		}
		r.current, r.next = r.next, r.current[:0]
	}
	for _, current := range r.current {
		value, ok := current.(string)
		if !ok {
			return nil, fmt.Errorf("expected string in resolved item, got %T", current)
		}
		r.out = append(r.out, value)
	}
	if len(r.out) == 0 {
		return nil, nil
	}
	return r.out, nil
}
