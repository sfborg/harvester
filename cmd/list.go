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
	"fmt"
	"log/slog"
	"sort"

	"github.com/fatih/color"
	"github.com/sfborg/harvester/internal/output"
	harvester "github.com/sfborg/harvester/pkg"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:           "list",
	Short:         "Shows list of supported datasets.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := []flagFunc{verboseFlag}
		for _, v := range flags {
			v(cmd)
		}

		cfg := config.New(opts...)

		hr := harvester.New(cfg)

		list := hr.List()
		slog.Info("show list of data sources")
		if cfg.WithVerbose {
			out := output.New(list)
			out.Table()
			fmt.Printf(
				"\n%s - require manual steps.\n",
				color.RedString("*"),
			)
		} else {
			var labels []string
			for k := range list {
				labels = append(labels, k)
			}
			sort.Strings(labels)
			for i, v := range labels {
				fmt.Printf("%0.2d %s", i+1, v)
				fmt.Println()
			}
			fmt.Printf(
				"\n%s - require manual steps. Use --verbose for details\n",
				color.RedString("*"),
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("verbose", "v", false, "Return details")
}
