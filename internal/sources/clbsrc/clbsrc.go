package clbsrc

import (
	"fmt"
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/internal/clb"
	"github.com/sfborg/harvester/internal/sysio"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type clbsrc struct {
	data.Convertor
	cfg       config.Config
	coldpPath string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "clb",
		Name:  "ChecklistBank",
		Notes: `Generic ChecklistBank export source.
Exports any ChecklistBank dataset as ColDP and converts to SFGA.
Requires --clb-dataset-id, --clb-user, and --clb-password flags.
Optionally accepts --clb-taxon-id to filter to a subtree.

Example:
  harvester get clb --clb-dataset-id 1027 \
    --clb-user USER --clb-password PASS \
    --clb-taxon-id 38829`,
		ManualSteps: false,
	}
	res := clbsrc{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (c *clbsrc) Download() (string, error) {
	if c.cfg.SkipDownload {
		return "", nil
	}

	if c.cfg.LoadFile != "" {
		slog.Info("using local file", "file", c.cfg.LoadFile)
		gn.Info("Using local file: %s", c.cfg.LoadFile)
		c.coldpPath = c.cfg.LoadFile
		return c.cfg.LoadFile, nil
	}

	if c.cfg.CLBDatasetID == 0 {
		return "", fmt.Errorf(
			"--clb-dataset-id is required for the clb source",
		)
	}
	if c.cfg.CLBUser == "" || c.cfg.CLBPassword == "" {
		return "", fmt.Errorf(
			"--clb-user and --clb-password are required for the clb source",
		)
	}

	err := sysio.ResetCache(c.cfg)
	if err != nil {
		return "", err
	}

	client := clb.New(c.cfg.CLBApi, c.cfg.CLBUser, c.cfg.CLBPassword)

	err = client.Login()
	if err != nil {
		return "", err
	}

	er := clb.ExportRequest{
		Format:         c.cfg.CLBFormat,
		Synonyms:       c.cfg.CLBSynonyms,
		BareNames:      c.cfg.CLBBareNames,
		Extended:       c.cfg.CLBExtended,
		Extinct:        c.cfg.CLBExtinct,
		Classification: c.cfg.CLBClassification,
		TaxGroups:      c.cfg.CLBTaxGroups,
		MinRank:        c.cfg.CLBMinRank,
		TabFormat:      c.cfg.CLBTabFormat,
	}

	if c.cfg.CLBTaxonID != "" {
		er.Root = &clb.Root{ID: c.cfg.CLBTaxonID}
	}

	exportKey, err := client.TriggerExport(c.cfg.CLBDatasetID, er)
	if err != nil {
		return "", err
	}

	downloadURL, err := client.WaitForExport(exportKey)
	if err != nil {
		return "", err
	}

	path, err := client.DownloadExport(downloadURL, c.cfg.DownloadDir)
	if err != nil {
		return "", err
	}

	c.coldpPath = path
	return path, nil
}

func (c *clbsrc) Extract(_ string) error {
	return nil
}

func (c *clbsrc) InitSfga() (sfga.Archive, error) {
	sysio.EmptyDir(c.cfg.SfgaDir)

	coldpPath := c.coldpPath
	if coldpPath == "" {
		var err error
		coldpPath, err = clb.FindColdpZip(c.cfg.DownloadDir)
		if err != nil {
			return nil, err
		}
	}

	return clb.ColdpToSfga(coldpPath, c.cfg)
}

func (c *clbsrc) ToSfga(_ sfga.Archive) error {
	return nil
}
