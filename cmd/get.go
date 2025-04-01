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
	"log/slog"
	"os"
	"strconv"

	harvester "github.com/sfborg/harvester/pkg"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <label-or-id> [sfga-output-path] [flags]",
	Short: "Converts registered source to SFGA file.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 || len(args) > 2 {
			cmd.Help()
			os.Exit(0)
		}

		flags := []flagFunc{
			skipFlag, fileFlag, zipFlag, delimFlag, quotesFlag, badRowFlag,
			dateFlag,
		}

		for _, v := range flags {
			v(cmd)
		}

		cfg := config.New(opts...)
		hr := harvester.New(cfg)

		l := getLabel(hr, args[0])
		if l == "" {
			slog.Error("Cannot find given source label", "input", args[0])
			slog.Info("Use `list` command to find registered sources")
			os.Exit(1)
		}

		outPath := l
		if len(args) == 2 {
			outPath = args[1]
		}

		err := hr.Convert(l, outPath)
		if err != nil {
			slog.Error(
				"Cannot convert source to SFGA file",
				"source", l, "path", outPath, "error", err,
			)
			os.Exit(1)
		}
	},
}

func getLabel(hr harvester.Harvester, ds string) string {
	list := hr.List()
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
	rootCmd.Flags().StringP(
		"delimiter", "d", "",
		"a delimiter for delimiter-separated files like CSV/TSV/PSV etc.",
	)
}
