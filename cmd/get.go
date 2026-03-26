/*
Copyright © 2025 Dmitry Mozzherin <dmozzherin@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/gnames/gn"
	"github.com/sfborg/harvester/internal/clb"
	"github.com/sfborg/harvester/internal/sources/clbsrc"
	harvester "github.com/sfborg/harvester/pkg"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:           "get <label-or-id> [sfga-output-path] [flags]",
	Short:         "Converts registered source to SFGA file.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			cmd.Help()
			return nil
		}

		flags := []flagFunc{
			skipFlag, fileFlag, zipFlag, delimFlag, quotesFlag, badRowFlag,
			dateFlag, dataVersionFlag, schemaFlag,
			clbApiFlag, clbUserFlag, clbPasswordFlag, clbDatasetIDFlag,
			clbDatasetAliasFlag, clbTaxonIDFlag,
			clbSynonymsFlag, clbBareNamesFlag,
			clbExtendedFlag, clbExtinctFlag, clbClassificationFlag,
			clbTaxGroupsFlag, clbMinRankFlag,
		}

		for _, v := range flags {
			v(cmd)
		}

		cfg := config.New(opts...)
		hr := harvester.New(cfg)

		l := getLabel(hr, args[0])
		if l == "" {
			err := fmt.Errorf("cannot find given source label for %s", args[0])
			gn.PrintErrorMessage(err)
			gn.Info("use `list` command to find registered sources")
			return err
		}

		// Prompt for missing CLB parameters if needed.
		if needsCLBPrompt(l, cfg) {
			cfg, hr = promptCLBParams(l, cfg)
		}

		outPath := l
		if len(args) == 2 {
			outPath = args[1]
		} else if cfg.CLBDatasetID != 0 {
			outPath = resolveCLBAlias(cfg)
		}

		err := hr.Get(l, outPath)
		if err != nil {
			fmt.Printf("err: %s", err)
			gn.PrintErrorMessage(err)
			return err
		}
		return nil
	},
}

// resolveCLBAlias returns the output file base name for a CLB dataset.
// It uses CLBDatasetAlias if set, otherwise fetches the alias from
// the CLB API metadata.
func resolveCLBAlias(cfg config.Config) string {
	if cfg.CLBDatasetAlias != "" {
		return cfg.CLBDatasetAlias
	}

	client := clb.New(cfg.CLBApi, "", "")
	alias, err := client.FetchDatasetAlias(cfg.CLBDatasetID)
	if err != nil {
		gn.Warn("Could not fetch dataset alias: %s", err)
		return fmt.Sprintf("clb-%d", cfg.CLBDatasetID)
	}

	return alias
}

// needsCLBPrompt returns true when the source needs interactive
// input for missing CLB parameters.
func needsCLBPrompt(label string, cfg config.Config) bool {
	if !clbsrc.IsCLBSource(label) {
		return false
	}
	if cfg.SkipDownload || cfg.LoadFile != "" {
		return false
	}
	if label == "clb" && cfg.CLBDatasetID == 0 {
		return true
	}
	return cfg.CLBUser == "" || cfg.CLBPassword == ""
}

// promptCLBParams asks the user for missing CLB parameters
// interactively. The password is hidden. It returns an updated
// config and harvester.
func promptCLBParams(
	label string, cfg config.Config,
) (config.Config, harvester.Harvester) {
	reader := bufio.NewReader(os.Stdin)

	if label == "clb" && cfg.CLBDatasetID == 0 {
		fmt.Print("ChecklistBank dataset ID: ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if id, err := strconv.Atoi(line); err == nil && id > 0 {
			opts = append(opts, config.OptCLBDatasetID(id))
		}
	}

	if cfg.CLBUser == "" {
		fmt.Print("ChecklistBank username: ")
		user, _ := reader.ReadString('\n')
		user = strings.TrimSpace(user)
		opts = append(opts, config.OptCLBUser(user))
	}

	if cfg.CLBPassword == "" {
		fmt.Print("ChecklistBank password: ")
		pw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err == nil {
			opts = append(opts, config.OptCLBPassword(string(pw)))
		}
	}

	cfg = config.New(opts...)
	hr := harvester.New(cfg)
	return cfg, hr
}

func getLabel(hr harvester.Harvester, ds string) string {
	var list []string
	for k := range hr.List() {
		list = append(list, k)
	}
	sort.Strings(list)
	idx, _ := strconv.Atoi(ds)
	if idx > 0 && len(list) >= idx {
		return list[idx-1]
	}
	for _, v := range list {
		if ds == v {
			return v
		}
	}
	return ""
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP(
		"file", "f", "", "get data from local file or URL",
	)
	getCmd.Flags().BoolP(
		"skip-download", "s", false, "skip downloading and extracting source",
	)
	getCmd.Flags().BoolP(
		"zip-output", "z", false, "compress output with zip",
	)
	getCmd.Flags().BoolP(
		"no-quotes", "Q", false,
		"for tsv, pipe-delimited without quotes for fields",
	)
	rootCmd.Flags().StringP(
		"wrong-fields-num", "w", "",
		`how to process rows with wrong fields number
     choices: 'stop', 'ignore', 'process'
     default: 'process'`,
	)
	getCmd.Flags().StringP(
		"issued-date", "d", "", "date when the archive was issued",
	)
	getCmd.Flags().StringP(
		"data-version", "v", "", "sets the version of the dataset",
	)
	rootCmd.Flags().StringP(
		"delimiter", "d", "",
		"a delimiter for delimiter-separated files like CSV/TSV/PSV etc.",
	)
	getCmd.Flags().StringP(
		"schema", "S", "",
		"path to local schema.sql file (instead of fetching from GitHub)",
	)

	// ChecklistBank API flags.
	getCmd.Flags().String(
		"clb-api", "",
		"ChecklistBank API URL (default https://api.checklistbank.org)",
	)
	getCmd.Flags().String("clb-user", "", "ChecklistBank username")
	getCmd.Flags().String("clb-password", "", "ChecklistBank password")
	getCmd.Flags().Int("clb-dataset-id", 0, "ChecklistBank dataset ID")
	getCmd.Flags().String(
		"clb-dataset-alias", "",
		"output file base name (default: from CLB dataset metadata)",
	)
	getCmd.Flags().String(
		"clb-taxon-id", "",
		"root taxon ID for CLB export filtering",
	)

	// ChecklistBank export option flags.
	getCmd.Flags().Bool("clb-synonyms", true, "include synonyms in CLB export")
	getCmd.Flags().Bool("clb-bare-names", false, "include bare names in CLB export")
	getCmd.Flags().Bool("clb-extended", true, "request extended CLB export")
	getCmd.Flags().Bool(
		"clb-extinct", false,
		"filter by extinction status in CLB export",
	)
	getCmd.Flags().Bool(
		"clb-classification", true,
		"include classification in CLB export",
	)
	getCmd.Flags().Bool(
		"clb-tax-groups", true,
		"include taxonomic groups in CLB export",
	)
	getCmd.Flags().String("clb-min-rank", "", "minimum rank for CLB export")
}
