package cmd

import (
	"os"

	"github.com/bmeg/jsonschemagraph/cmd/gengraph"
	"github.com/bmeg/jsonschemagraph/cmd/schema_graph"
	"github.com/bmeg/jsonschemagraph/cmd/schema_lint"
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "jsonschemagraph",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(gengraph.Cmd)
	RootCmd.AddCommand(schema_lint.Cmd)
	RootCmd.AddCommand(schema_graph.Cmd)

}

var genBashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions file",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
