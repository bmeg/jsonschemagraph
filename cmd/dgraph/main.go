package dgraph

import (
	"fmt"
	"log"

	"github.com/bmeg/jsonschemagraph/compile"
	"github.com/bmeg/jsonschemagraph/graph"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dgraph [schema dir]",
	Short: "Generates a d2 file to visualize graph schema structure",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		sch, _ := graph.Load(args[0])
		log.Printf("digraph {\n")
		for _, cls := range sch.Classes {
			if cls.Title != "" {
				log.Printf("\t%s\n", cls.Title)
			}
		}
		// Not sure if the print statements here are correct the but output graph seems reasonable
		for _, cls := range sch.Classes {
			if len(cls.Extensions) > 0 {
				gExt := cls.Extensions[0].(*compile.HyperMediaExt)
				for _, v := range gExt.Targets {
					fmt.Printf("\t%s -> %s: %s\n", cls.Title, v.Schema.Title, v.Rel)
					if v.TargetHints.Backref != nil {
						fmt.Printf("\t%s -> %s: %s\n", v.Schema.Title, cls.Title, v.TargetHints.Backref[0])
					}
				}
			}
		}
		log.Printf("}\n")
		return nil
	},
}
