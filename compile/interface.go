package compile

import (
	"github.com/bmeg/jsonschema/v6"
)

var GExtUrl = "graphExtMeta.json"

var GraphExtMeta = []byte(`{
	"properties": {
		"anchor": {
			"type": "string",
			"format": "uri-template"
		},
		"anchorPointer": {
			"type": "string",
			"anyOf": [
				{
					"format": "json-pointer"
				},
				{
					"format": "relative-json-pointer"
				}
			]
		},
		"rel": {
			"anyOf": [
				{
					"type": "string"
				},
				{
					"type": "array",
					"items": {
						"type": "string"
					},
					"minItems": 1
				}
			]
		},
		"href": {
			"type": "string",
			"format": "uri-template"
		},
		"templatePointers": {
			"type": "object",
			"additionalProperties": {
				"type": "string",
				"anyOf": [
					{
						"format": "json-pointer"
					},
					{
						"format": "relative-json-pointer"
					}
				]
			}
		},
		"templateRequired": {
			"type": "array",
			"items": {
				"type": "string"
			},
			"uniqueItems": true
		},
		"title": {
			"type": "string"
		},
		"description": {
			"type": "string"
		},
		"$comment": {
			"type": "string"
		}
	}
}`)

type HyperMediaExt struct {
	Targets []Target
}

type Target struct {
	Schema           *jsonschema.Schema `json:"schema,omitempty"`
	Href             string             `json:"href,omitempty"`
	Rel              string             `json:"rel,omitempty"`
	TargetHints      TargetHints        `json:"targetHints,omitempty"`
	TargetSchema     TargetSchema       `json:"targetSchema,omitempty"`
	TemplatePointers TemplatePointers   `json:"templatePointers,omitempty"`
	TemplateRequired []string           `json:"templateRequired,omitempty"`
}

type TargetHints struct {
	Backref     []string `json:"backref,omitempty"`
	Direction   []string `json:"direction,omitempty"`
	Multiplicty []string `json:"multiplicty,omitempty"`
	RegexMatch  []string `json:"regex_match,omitempty"`
}

type TargetSchema struct {
	Ref string `json:"ref,omitempty"`
}

type TemplatePointers struct {
	Id string `json:"id,omitempty"`
	// This field is meant to be "pre-split" and popualted during compile time
	// as remove repeated execution split operations
	SplittedId []string `json:"splitted_id,omitempty"`
}
