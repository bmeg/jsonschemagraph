package cmd

import (
	"github.com/bmeg/jsonschemagraph/cmd/generate"
	"github.com/bmeg/jsonschemagraph/cmd/graphql"
	"github.com/bmeg/jsonschemagraph/cmd/lintSchema"
	"github.com/bmeg/jsonschemagraph/cmd/validate"

	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "jsonschemagraph",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(validate.Cmd)
	RootCmd.AddCommand(lintSchema.Cmd)
	RootCmd.AddCommand(generate.Cmd)
	RootCmd.AddCommand(graphql.Cmd)
}
