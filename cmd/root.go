package cmd

import (
	"os"

	"github.com/bmeg/jsonschemagraph/cmd/data_validate"
	"github.com/bmeg/jsonschemagraph/cmd/gen_dir"
	"github.com/bmeg/jsonschemagraph/cmd/gen_graph"
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
	RootCmd.AddCommand(gen_dir.Cmd)
	RootCmd.AddCommand(gen_graph.Cmd)
	RootCmd.AddCommand(schema_lint.Cmd)
	RootCmd.AddCommand(schema_graph.Cmd)
	RootCmd.AddCommand(data_validate.Cmd)

}

var genBashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions file",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
