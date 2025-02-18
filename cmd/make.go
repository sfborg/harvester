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

	"github.com/sfborg/harvester/internal/io/sysio"
	harvester "github.com/sfborg/harvester/pkg"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/spf13/cobra"
)

// makeCmd represents the make command
var makeCmd = &cobra.Command{
	Use:   "make <label-or-id> [sfga-output-path] [flags]",
	Short: "Converts registered source to SFGA file.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 || len(args) > 2 {
			cmd.Help()
			os.Exit(0)
		}

		cfg := config.New(opts...)

		err := sysio.ResetCache(cfg)
		if err != nil {
			slog.Error("Cannot reset cache", "error", err)
			os.Exit(1)
		}

		hr := harvester.New(cfg)

		ds := getDataSource(hr, args[0])
		if ds == "" {
			slog.Error("Cannot find given source label", "input", args[0])
			slog.Info("Use `list` command to find registered sources")
			os.Exit(1)
		}

		path := ds
		if len(args) == 2 {
			path = args[1]
		}

		err = hr.Convert(ds, path)
		if err != nil {
			slog.Error(
				"Cannot convert source to SFGA file",
				"source", ds, "path", path, "error", err,
			)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(makeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// makeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// makeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getDataSource(hr harvester.Harvester, ds string) string {
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
