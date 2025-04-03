package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gnames/gnfmt"
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

func dateFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("issued-date")
	if s != "" {
		opts = append(opts, config.OptArchiveDate(s))
	}
}

func dataVersionFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("data-version")
	if s != "" {
		opts = append(opts, config.OptArchiveVersion(s))
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

func badRowFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("wrong-fields-num")
	switch s {
	case "":
		return
	case "stop":
		opts = append(opts, config.OptBadRow(gnfmt.ErrorBadRow))
	case "ignore":
		opts = append(opts, config.OptBadRow(gnfmt.SkipBadRow))
	case "process":
		opts = append(opts, config.OptBadRow(gnfmt.ProcessBadRow))
	default:
		slog.Warn("Unknown setting for wrong-fields-num, keeping default",
			"setting", s)
		slog.Info("Supported values are: 'stop', 'ignore', 'process' (default)")
	}
}

func quotesFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("no-quotes")
	if b {
		opts = append(opts, config.OptWithoutQuotes(true))
	}
}

func delimFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("delimiter")
	switch s {
	case "\t", ",", "|":
		opts = append(opts, config.OptColSep(s))
	default:
		slog.Warn(
			"Unknown delimiter, using automatic detection.", "delimiter", s,
		)
	}
}
