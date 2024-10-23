package data_validate

import (
	"encoding/json"
	"strings"
	"log"

	"github.com/bmeg/golib"
	"github.com/bmeg/jsonschema/v5"
	"github.com/bmeg/jsonschemagraph/graph"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "data-validate [schemaFile] [inputFile]",
	Short: "Data Validate",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		schemaFile := args[0]
		inputPath := args[1]

		jsonschema.Loaders["file"] = graph.YamlLoader

		compiler := jsonschema.NewCompiler()
		compiler.ExtractAnnotations = true

		sch, err := compiler.Compile(schemaFile)
		if err != nil {
			log.Fatalf("Error compiling %s : %s\n", schemaFile, err)
		} else {
			if len(sch.Types) == 1 && sch.Types[0] == "object" {
				log.Printf("OK: %s %s (%s)\n", schemaFile, sch.Title, sch.Title)
			}
		}

		var reader chan []byte
		if strings.HasSuffix(inputPath, ".gz") {
			reader, err = golib.ReadGzipLines(inputPath)
		} else {
			reader, err = golib.ReadFileLines(inputPath)
		}
		if err != nil {
			return err
		}

		procChan := make(chan map[string]interface{}, 100)
		go func() {
			for line := range reader {
				o := map[string]interface{}{}
				if len(line) > 0 {
					json.Unmarshal(line, &o)
					procChan <- o
				}
			}
			close(procChan)
		}()

		validCount := 0
		errorCount := 0
		for row := range procChan {
			err = sch.Validate(row)
			if err != nil {
				errorCount++
				log.Printf("Error: %s\n", err)
			} else {
				validCount++
			}
		}
		log.Printf("%s results: %d valid records %d invalid records\n", inputPath, validCount, errorCount)
		return nil
	},
}
