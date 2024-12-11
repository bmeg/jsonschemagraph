package gen_graphql

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/bmeg/jsonschemagraph/schconv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var jsonSchemaFile string
var yamlSchemaDir string
var graphName string
var configPath string
var writeFile bool = false

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
			data, err := ioutil.ReadFile(configPath)
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
			_, err := schconv.ParseGraphFile(jsonSchemaFile, "jsonSchema", graphName, config.DependencyOrder, writeFile)
			if err != nil {
				return err
			}
		}
		if yamlSchemaDir != "" && graphName != "" {
			log.Printf("Loading Yaml Schema dir: %s", yamlSchemaDir)
			_, err := schconv.ParseGraphFile(yamlSchemaDir, "yamlSchema", graphName, config.DependencyOrder, writeFile)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	gqlflags := Cmd.Flags()
	gqlflags.BoolVar(&writeFile, "writeFile", false, "Write file to disk")
	gqlflags.StringVar(&jsonSchemaFile, "jsonSchema", "", "Json Schema")
	gqlflags.StringVar(&yamlSchemaDir, "yamlSchemaDir", "", "Name of YAML schemas dir")
	gqlflags.StringVar(&configPath, "configPath", "", "Path of Config file for determining the subset of ")
	gqlflags.StringVar(&graphName, "graphName", "", "Name of schemaGraph")
}
