package schema_graph

import (
	"fmt"

	jsgraph "github.com/bmeg/jsonschemagraph/util"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "schema-graph [schema dir]",
	Short: "Generates a d2 file to visualize graph schema structure",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		var Node_Types = [...]string {"Practitioner", "ReserachStudy",
		"Patient", "ResearchSubject", "Substance", "Specimen", "Observation", "Condition",
		"Procedure", "DocumentReference", "Task", "ImagingStudy", "Encounter", "Allele",
		"AlleleEffect", "CopyNumberAlteration", "DrugResponse", "Exon", "Gene", "GeneExpression",
		"GenePhenotypeAssociation", "Interaction", "Methylation", "Phenotype", "Protein", "ProteinCompoundAssociation",
		"ProteinStructure", "SomaticCallset", "SomaticVariant", "Transcript"}

		sch, _ := jsgraph.Load(args[0])
		fmt.Printf("digraph {\n")
		var NodeFound bool
		for _, cls := range sch.Classes {
			for _, node := range Node_Types {
				NodeFound = false
				//fmt.Println("cls.Title", cls.Title)
				if node == cls.Title {
					NodeFound = true
					break
				}
			}
			if cls.Title != "" && NodeFound == true {
				fmt.Printf("\t%s\n", cls.Title)
			}
		}
		// Not sure if the print statements here are correct the but output graph seems reasonable
		// filter by set of node types:
		//
		for _, cls := range sch.Classes {
			if ext, ok := cls.Extensions[jsgraph.GraphExtensionTag]; ok {
				gExt := ext.(jsgraph.GraphExtension)
				for _, v := range gExt.Targets {
					schemaFound, TitleFound, AllFound := false, false, false
					for _, node := range Node_Types {
						if node == v.Schema.Title {
							schemaFound = true
						}
						if node == cls.Title {
							TitleFound = true
						}
						if schemaFound && TitleFound{
							AllFound = true
							break
						}
					}
					if AllFound{
						fmt.Printf("\t%s -> %s\n", cls.Title, v.Schema.Title)
						/*if v.Backref != "" {
							fmt.Printf("\t%s -> %s: %s\n", v.Schema.Title, cls.Title, v.Backref)
						}*/
					}
				}

			}
		}

		fmt.Printf("}\n")
		return nil
	},
}

func init() {
}
