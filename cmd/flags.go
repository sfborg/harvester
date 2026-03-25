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

func clbApiFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("clb-api")
	if s != "" {
		opts = append(opts, config.OptCLBApi(s))
	}
}

func clbUserFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("clb-user")
	if s != "" {
		opts = append(opts, config.OptCLBUser(s))
	}
}

func clbPasswordFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("clb-password")
	if s != "" {
		opts = append(opts, config.OptCLBPassword(s))
	}
}

func clbDatasetIDFlag(cmd *cobra.Command) {
	i, _ := cmd.Flags().GetInt("clb-dataset-id")
	if i != 0 {
		opts = append(opts, config.OptCLBDatasetID(i))
	}
}

func clbDatasetAliasFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("clb-dataset-alias")
	if s != "" {
		opts = append(opts, config.OptCLBDatasetAlias(s))
	}
}

func clbTaxonIDFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("clb-taxon-id")
	if s != "" {
		opts = append(opts, config.OptCLBTaxonID(s))
	}
}

func clbSynonymsFlag(cmd *cobra.Command) {
	if cmd.Flags().Changed("clb-synonyms") {
		b, _ := cmd.Flags().GetBool("clb-synonyms")
		opts = append(opts, config.OptCLBSynonyms(b))
	}
}

func clbBareNamesFlag(cmd *cobra.Command) {
	if cmd.Flags().Changed("clb-bare-names") {
		b, _ := cmd.Flags().GetBool("clb-bare-names")
		opts = append(opts, config.OptCLBBareNames(b))
	}
}

func clbExtendedFlag(cmd *cobra.Command) {
	if cmd.Flags().Changed("clb-extended") {
		b, _ := cmd.Flags().GetBool("clb-extended")
		opts = append(opts, config.OptCLBExtended(b))
	}
}

func clbExtinctFlag(cmd *cobra.Command) {
	if cmd.Flags().Changed("clb-extinct") {
		b, _ := cmd.Flags().GetBool("clb-extinct")
		opts = append(opts, config.OptCLBExtinct(&b))
	}
}

func clbClassificationFlag(cmd *cobra.Command) {
	if cmd.Flags().Changed("clb-classification") {
		b, _ := cmd.Flags().GetBool("clb-classification")
		opts = append(opts, config.OptCLBClassification(b))
	}
}

func clbTaxGroupsFlag(cmd *cobra.Command) {
	if cmd.Flags().Changed("clb-tax-groups") {
		b, _ := cmd.Flags().GetBool("clb-tax-groups")
		opts = append(opts, config.OptCLBTaxGroups(b))
	}
}

func clbMinRankFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("clb-min-rank")
	if s != "" {
		opts = append(opts, config.OptCLBMinRank(s))
	}
}

