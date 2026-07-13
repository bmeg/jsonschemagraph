package graph

import (
	"fmt"
	"log"

	"github.com/bmeg/jsonschema/v6"
	"github.com/bmeg/jsonschemagraph/compile"
)

type GraphSchema struct {
	Classes   map[string]*jsonschema.Schema
	EdgePlans map[string]*ClassEdgePlan
	Compiler  *jsonschema.Compiler
}

type ClassEdgePlan struct {
	EstimatedEdges int
	Rules          []EdgeRule
}

type EdgeRule struct {
	Rel           string
	Backref       string
	HasBackref    bool
	MatchPrefix   string
	AllowAnyMatch bool
	PointerKey    string
	Path          []string
}

func (s *GraphSchema) Validate(classID string, data map[string]any) error {
	class := s.GetClass(classID)
	if class != nil {
		return class.Validate(data)
	}
	return fmt.Errorf("class '%s' not found", classID)
}

func (s *GraphSchema) ValidateFast(classID string, data map[string]any) error {
	class := s.GetClass(classID)
	if class != nil {
		return class.ValidateFast(data)
	}
	return fmt.Errorf("class '%s' not found", classID)
}

func (s *GraphSchema) ListClasses() []string {
	out := []string{}
	for c := range s.Classes {
		out = append(out, c)
	}
	return out
}

func (s *GraphSchema) GetClass(classID string) *jsonschema.Schema {
	if class, ok := s.Classes[classID]; ok {
		return class
	}

	sch, err := s.Compiler.Compile(classID)
	if err != nil {
		log.Printf("compile error: %s", err)
		return nil
	}
	if sch.Title != "" {
		s.Classes[sch.Title] = sch
		s.EdgePlans[sch.Title] = compileEdgePlan(sch)
	}
	return sch
}

func (s *GraphSchema) GetEdgePlan(classID string) *ClassEdgePlan {
	if plan, ok := s.EdgePlans[classID]; ok {
		return plan
	}
	class := s.GetClass(classID)
	if class == nil {
		return nil
	}
	plan := compileEdgePlan(class)
	s.EdgePlans[classID] = plan
	return plan
}

func compileEdgePlan(class *jsonschema.Schema) *ClassEdgePlan {
	if class == nil || len(class.Extensions) == 0 {
		return &ClassEdgePlan{}
	}
	ext, ok := class.Extensions[0].(*compile.HyperMediaExt)
	if !ok || ext == nil {
		return &ClassEdgePlan{}
	}
	plan := &ClassEdgePlan{
		EstimatedEdges: len(ext.Targets) * 2,
		Rules:          make([]EdgeRule, 0, len(ext.Targets)),
	}
	for _, target := range ext.Targets {
		rule := EdgeRule{
			Rel:        target.Rel,
			PointerKey: target.TemplatePointers.Id,
			Path:       target.TemplatePointers.SplittedId,
		}
		if len(target.TargetHints.Backref) > 0 {
			rule.Backref = target.TargetHints.Backref[0]
			rule.HasBackref = rule.Backref != ""
		}
		if len(target.TargetHints.RegexMatch) > 0 {
			rule.MatchPrefix = target.TargetHints.RegexMatch[0]
			rule.AllowAnyMatch = rule.MatchPrefix == "Resource/*"
		}
		plan.Rules = append(plan.Rules, rule)
	}
	return plan
}
