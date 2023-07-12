## Setup

To run:
```
go run cmds/gengraph/main.go
```

To Test:
```
cd cmds/gengraph
go test
```
Since some of the source data has keys that have null values some of these test cases will fail. This is expected. 

Schemas are rc7 schemas with the new schema format

Note: Some changes had to be made from the generated schemas since some of the links were referencing schemas that didn't exist with their $ref keys. The correct reference schemas were taken from the old schemas.


Expected Program structure:
```bash
├── Makefile
├── README.md
├── cmds
│   └── gengraph
│       ├── main_test.go
│       └── main.go
├── data
│   ├── ensembl_exon.json.gz
│   ├── ensembl_gene.json.gz
│   └── ensembl_transcript.json.gz
├── generate.go
├── go.mod
├── go.sum
├── loader.go
├── methods.go
├── output
├── schema_extension_definition.json
├── bmeg_schemas
│   ├── exon.yaml
│   ├── gene.yaml
│   ├──transcript.yaml ... all thirty schemas
└── validate.go
```