## Setup

To build, install from root:

```
go build . ; go install .
```

benchmark new vs old:

```
jsonschemagraph generate iceberg/schemas/graph/graph-fhir.json  OUT2  76.73s user 7.21s system 178% cpu 47.103 total

vs

jsonschemagraph gen-dir iceberg/schemas/graph/graph-fhir.json  ../../g3t_projs/dev-cbds/cbds-htan/META OUT  175.85s user 14.78s system 176% cpu 1:47.74 total
```

```
Usage:
  jsonschemagraph [command]

Available Commands:
  completion    Generate the autocompletion script for the specified shell
  data-validate Data Validate
  generate      Generates edges and vertices from source data files and schemas
  graphql       generate graphql schemas
  help          Help about any command
  schema-lint   Checks a directory of yaml schemas for syntax errors

Flags:
  -h, --help   help for jsonschemagraph

Use "jsonschemagraph [command] --help" for more information about a command
```

To install d2 graph description language used for schema-graph command:

```
curl -fsSL https://d2lang.com/install.sh | sh -s --
```

Since some of the source data has keys that have null values some of these test cases will fail. This is expected.

### Example use case adding Namespace DNS and Auth_Resource_Path, Caliper specific data fields:

specify custom namespace for edge generation with "namespace" key in extra args, specify gen3 authentication ProjectId with
/programs/{program}/projects/{project} string template format. Ex:

```
jsonschemagraph generate ../iceberg/schemas/graph DATA OUT --extraArgs '{"auth_resource_path": "/programs/ohsu/projects/test", "namespace": "CALIPERIDP.org"}'
```

### Example commands for generating graphql schema from jsonschema

```
jsonschemagraph graphql --graphName CALIPER --jsonSchema graph-fhir.json  --configPath config.yaml --writeIntermediateFile
```

This translator does not support codeable references
