package cmd

import (
	"fmt"
	"os"

	"github.com/gnames/gnparser/ent/nomcode"
	harvester "github.com/sfborg/harvester/pkg"
	"github.com/sfborg/harvester/pkg/config"
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

func verboseFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("verbose")
	if b {
		opts = append(opts, config.OptWithVerbose(true))
	}
}

func zipFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("zip-output")
	if b {
		opts = append(opts, config.OptWithZipOutput(true))
	}
}

func fileFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("file")
	if s != "" {
		opts = append(opts, config.OptLocalFile(s))
	}
}

func skipFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("skip-download")
	if b {
		opts = append(opts, config.OptSkipDownload(true))
	}
}

func codeFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("code")
	if s != "" {
		code := nomcode.New(s)
		opts = append(opts, config.OptCode(code))
	}
}
