package gen_graphql

import (
	"fmt"
	"log"
	"os"

	schema "github.com/bmeg/jsonschemagraph/graphql"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var jsonSchemaFile string
var yamlSchemaDir string
var graphName string
var configPath string
var writeIntermediateFile bool = false

type Config struct {
	DependencyOrder []string `yaml:"dependency_order"`
}

var Cmd = &cobra.Command{
	Use:   "gen-graphql",
	Short: "Load graph schemas",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonSchemaFile == "" && yamlSchemaDir == "" {
			return fmt.Errorf("No schema file was provided")
		}

		config := Config{DependencyOrder: []string{}}
		if configPath != "" {
			data, err := os.ReadFile(configPath)
			if err != nil {
				log.Fatalf("Failed to read YAML file: %v", err)
			}
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				log.Fatalf("Failed to parse YAML file: %v", err)
			}
		} else {
			fmt.Printf("Warning: No config file was provided, all vertices will be rendered has queries")
		}

		if jsonSchemaFile != "" && graphName != "" {
			log.Printf("Loading Json Schema file: %s", jsonSchemaFile)
			graphs, err := schema.ParseGraphFile(jsonSchemaFile, "jsonSchema", graphName, config.DependencyOrder, writeIntermediateFile)
			if err != nil {
				return err
			}
			_ = schema.GripGraphqltoGraphql(graphs[0])
		}
		if yamlSchemaDir != "" && graphName != "" {
			log.Printf("Loading Yaml Schema dir: %s", yamlSchemaDir)
			graphs, err := schema.ParseGraphFile(yamlSchemaDir, "yamlSchema", graphName, config.DependencyOrder, writeIntermediateFile)
			if err != nil {
				return err
			}
			_ = schema.GripGraphqltoGraphql(graphs[0])
		}

		return nil
	},
}

func init() {
	gqlflags := Cmd.Flags()
	gqlflags.BoolVar(&writeIntermediateFile, "writeIntermediateFile", false, "Write writeIntermediateFile file to disk")
	gqlflags.StringVar(&jsonSchemaFile, "jsonSchema", "", "Json Schema")
	gqlflags.StringVar(&yamlSchemaDir, "yamlSchemaDir", "", "Name of YAML schemas dir")
	gqlflags.StringVar(&configPath, "configPath", "", "Path of Config file for determining the subset of ")
	gqlflags.StringVar(&graphName, "graphName", "", "Name of schemaGraph")
}
