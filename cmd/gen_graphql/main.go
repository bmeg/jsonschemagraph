package gen_graphql

import (
	"fmt"
	"log"

	"github.com/bmeg/jsonschemagraph/schconv"
	"github.com/spf13/cobra"
)

var jsonSchemaFile string
var yamlSchemaDir string
var graphName string

// https://github.com/bmeg/sifter/blob/51a67b0de852e429d30b9371d9975dbefe3a8df9/transform/graph_build.go#L86

var Cmd = &cobra.Command{
	Use:   "gen_graphql",
	Short: "Load graph schemas",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonSchemaFile == "" && yamlSchemaDir == "" {
			return fmt.Errorf("no schema file was provided")
		}

		if jsonSchemaFile != "" && graphName != "" {
			log.Printf("Loading Json Schema file: %s", jsonSchemaFile)
			_, err := schconv.ParseGraphFile(jsonSchemaFile, "jsonSchema", graphName)
			if err != nil {
				return err
			}

		}
		if yamlSchemaDir != "" && graphName != "" {
			log.Printf("Loading Yaml Schema dir: %s", yamlSchemaDir)
			_, err := schconv.ParseGraphFile(yamlSchemaDir, "yamlSchema", graphName)
			if err != nil {
				log.Println("HELLO ERROR HERE: ", err)
				return err
			}

		}
		return nil
	},
}

func init() {
	gqlflags := Cmd.Flags()
	gqlflags.StringVar(&jsonSchemaFile, "jsonSchema", "", "Json Schema")
	gqlflags.StringVar(&yamlSchemaDir, "yamlSchemaDir", "", "Name of YAML schemas dir")
	gqlflags.StringVar(&graphName, "graphName", "", "Name of schemaGraph")
}
