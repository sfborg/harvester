package cmd

import (
	"fmt"
	"log/slog"

	"github.com/gnames/gn"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/nomcode"
	harvester "github.com/sfborg/harvester/pkg"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/spf13/cobra"
)

type flagFunc func(cmd *cobra.Command)

func versionFlag(cmd *cobra.Command) bool {
	b, _ := cmd.Flags().GetBool("version")
	if b {
		version := harvester.GetVersion()
		fmt.Printf("\nVersion: %s\nBuild:   %s\n", version.Version, version.Build)
	}
	return b
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
		slog.Warn("unknown setting for wrong-fields-num, keeping default",
			"setting", s)
		gn.Warn("unknown <em>wrong-fields-num</em> setting <em>%s</em>", s)
		gn.Info("Supported values are: 'stop', 'ignore', 'process' (default)")
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
	case "":
		return
	case "\t", ",", "|":
		opts = append(opts, config.OptColSep(s))
	default:
		slog.Warn(
			"unknown delimiter, using automatic detection.", "delimiter", s,
		)
	}
}

func schemaFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("schema")
	if s != "" {
		opts = append(opts, config.OptLocalSchemaPath(s))
	}
}
