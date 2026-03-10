package output

import (
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sfborg/harvester/pkg/data"
)

type Output struct {
	OutputDatasets []OutputDataset
}

type OutputDataset struct {
	Index int
	Label string
	Title string
	Notes string
}

func New(list map[string]data.Convertor) *Output {
	res := make([]OutputDataset, len(list))
	var labels []string
	for k := range list {
		labels = append(labels, k)
	}
	sort.Strings(labels)
	for i, v := range labels {
		datum := list[v]
		label := datum.Label()
		if datum.ManualSteps() {
			label += color.RedString("*")
		}
		od := OutputDataset{
			Index: i + 1,
			Label: label,
			Title: datum.Name(),
			Notes: datum.Description(),
		}
		res[i] = od
	}
	return &Output{OutputDatasets: res}
}

func (o *Output) Table() {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Title", WidthMax: 20},
		{Name: "Notes"},
	})
	header := table.Row{"#", "Label", "Title", "Notes"}
	t.AppendHeader(header)

	for _, v := range o.OutputDatasets {
		row := table.Row{
			v.Index,
			v.Label,
			v.Title,
			v.Notes,
		}
		t.AppendRow(row)
		t.AppendSeparator()
	}
	t.Render()
}
