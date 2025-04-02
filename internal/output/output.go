package output

import (
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
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
		od := OutputDataset{
			Index: i + 1,
			Label: datum.Label(),
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
		{Name: "Written", Align: text.AlignRight},
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
	}
	t.Render()
}
