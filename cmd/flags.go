package cmd

import (
	"fmt"
	"os"

	harvester "github.com/sfborg/harvester/pkg"
	"github.com/spf13/cobra"
)

type flagFunc func(cmd *cobra.Command)

func versionFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("version")
	if b {
		version := harvester.GetVersion()
		fmt.Print(version)
		os.Exit(0)
	}
}
