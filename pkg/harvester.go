package harvester

import (
	"fmt"
	"sort"

	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/list"
	"github.com/sfborg/harvester/pkg/config"
)

type harvester struct {
	cfg config.Config
}

func New(cfg config.Config) Harvester {
	res := harvester{cfg: cfg}
	return &res
}

func (h *harvester) List() []string {
	var res []string
	for k := range list.DataSets {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (h *harvester) Convert(label string) error {
	var err error
	var ds data.Convertor
	var ok bool
	if ds, ok = list.DataSets[label]; !ok {
		err = fmt.Errorf("Label '%s' does not exist", label)
		return err
	}
	err = ds.Download()
	if err != nil {
		return err
	}
	err = ds.ToSFGA()
	if err != nil {
		return err
	}
	return nil
}
