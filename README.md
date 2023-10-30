## Setup

To install, from root:
```
go install .
```

```
jsonschemagraph
Usage:
  jsonschemagraph [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  gen-graph    Gen graph
  help         Help about any command
  schema-graph Schema graph <dir>
  schema-lint  Schema lint

Flags:
  -h, --help   help for jsonschemagraph

Use "jsonschemagraph [command] --help" for more information about a command.
```

To test:
```
cd cmds/gengraph
go test
```

To install d2 graph description language used for schema-graph command:
```
curl -fsSL https://d2lang.com/install.sh | sh -s --
```

Since some of the source data has keys that have null values some of these test cases will fail. This is expected.

## Expected Program Structure:
```bash
├── README.md
├── cmd
│   ├── gengraph
│   │   ├── main.go
│   │   └── main_test.go
│   ├── root.go
│   └── schema_lint
│       └── main.go
├── go.mod
├── go.sum
├── main.go
├── data/ 
├── schemas/
├── output/
└── util
    ├── delete_empty.go
    ├── generate.go
    ├── loader.go
    ├── methods.go
    └── validate.go
```

Data is where your data files that you want to generate edges and vertices with will go. These files must be file type .gz

Output is an example name of the directory that the edges and vertices will be output to. You can specifiy whatever directory you want in the below gengraph command and the directory path will be created for you if it does not exist.

Schemas is the location of your schema files

## Example Commands
Generate edge and vertex files 
```
jsonschemagraph gen-graph [schema_directory_location] [data_file_location] [output_directory_location] [schema_class_name]
```
SWAPI Example: jsonschemagraph gen-graph schemas/ data/swapi/swapi_character.json.gz output character

BMEG Example: jsonschemagraph gen-graph schemas/bmeg_schemas data/ensembl_transcript.json.gz output Transcript

FHIR Example: jsonschemagraph gen-graph schemas/all_fhir_schemas aced-data/actual_fhir/Observation.json.gz out Observation

Note: the fhir data used in the above example is an ndjson file of FHIR Observations

Check to see if the schemas in a directory are valid
```
jsonschemagraph schema-lint [schema_directory_location]
```

Generate a d2 graphical representation of a directory of graph schemas
```
jsonschemagraph schema-graph [schema_directory_location] > in.d2
d2 --watch in.d2 out.svg 
```

