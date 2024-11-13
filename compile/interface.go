package compile

import "github.com/bmeg/jsonschema/v5"

var GraphExtensionTag = "json_schema_graph"

var GraphExtMeta = jsonschema.MustCompileString("graphExtMeta.json", `{"properties": {
	"anchor": {
		"type": "string",
		"format": "uri-template"
	},
	"anchorPointer": {
		"type": "string",
		"anyOf": [
			{ "format": "json-pointer" },
			{ "format": "relative-json-pointer" }
		]
	},
	"rel": {
		"anyOf": [
			{ "type": "string" },
			{
				"type": "array",
				"items": { "type": "string" },
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
				{ "format": "json-pointer" },
				{ "format": "relative-json-pointer" }
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

type GraphExtCompiler struct{}

type GraphExtension struct {
	Targets []Target
}

type Target struct {
	Schema           *jsonschema.Schema `json:"schema"`
	Href             string             `json:"href"`
	Rel              string             `json:"rel"`
	TargetHints      TargetHints        `json:"targetHints"`
	TargetSchema     TargetSchema       `json:"targetSchema"`
	TemplatePointers TemplatePointers   `json:"templatePointers"`
	TemplateRequired []string           `json:"templateRequired"`
}

type TargetHints struct {
	Backref     []string `json:"backref"`
	Direction   []string `json:"direction"`
	Multiplicty []string `json:"multiplicty"`
	RegexMatch  []string `json:"regex_match"`
}

type TargetSchema struct {
	Ref string `json:"ref"`
}

type TemplatePointers struct {
	Id string `json:"id"`
}
