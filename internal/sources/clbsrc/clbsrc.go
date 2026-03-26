// Package clbsrc provides a reusable ChecklistBank source
// implementation. Individual CLB dataset sources (e.g. wsc, scarabs)
// call New() with their dataset definition. Post-processing can be
// added by embedding the returned Convertor and overriding ToSfga.
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

// Dataset defines a ChecklistBank dataset for use with New().
type Dataset struct {
	// Label is the short identifier used in `harvester get <label>`.
	Label string

	// Name is the human-readable dataset name.
	Name string

	// Notes is the description shown in `harvester list -v`.
	Notes string

	// DatasetID is the ChecklistBank dataset key.
	// Zero means the user must supply --clb-dataset-id.
	DatasetID int

	// TaxonName is the default root taxon resolved by name search.
	// Empty means no default root taxon filter.
	TaxonName string

	// TaxonRank narrows the taxon name search (e.g. "class").
	TaxonRank string
}

// NewGeneric creates the generic "clb" source that requires
// --clb-dataset-id from the user.
func NewGeneric(cfg config.Config) data.Convertor {
	return New(cfg, Dataset{
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
	})
}

// New creates a data.Convertor backed by the ChecklistBank API.
// Sources that need post-processing can embed the returned Convertor
// and override ToSfga.
func New(cfg config.Config, ds Dataset) data.Convertor {
	set := data.DataSet{
		Label:       ds.Label,
		Name:        ds.Name,
		Notes:       ds.Notes,
		CLBSource:   true,
		ManualSteps: false,
	}
	return &clbSource{
		cfg:       cfg,
		ds:        ds,
		Convertor: base.New(cfg, &set),
	}
}

type clbSource struct {
	data.Convertor
	cfg       config.Config
	ds        Dataset
	coldpPath string
}

func (c *clbSource) Download() (string, error) {
	if c.cfg.SkipDownload {
		return "", nil
	}

	if c.cfg.LoadFile != "" {
		slog.Info("using local file", "file", c.cfg.LoadFile)
		gn.Info("Using local file: %s", c.cfg.LoadFile)
		c.coldpPath = c.cfg.LoadFile
		return c.cfg.LoadFile, nil
	}

	datasetID := c.ds.DatasetID
	if datasetID == 0 {
		datasetID = c.cfg.CLBDatasetID
	}
	if datasetID == 0 {
		return "", fmt.Errorf(
			"--clb-dataset-id is required for the clb source",
		)
	}

	if c.cfg.CLBUser == "" || c.cfg.CLBPassword == "" {
		if !clb.HasValidToken(c.cfg.CLBApi) {
			return "", fmt.Errorf(
				"--clb-user and --clb-password are required",
			)
		}
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

	taxonID, err := c.resolveTaxonID(client, datasetID)
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

	if taxonID != "" {
		er.Root = &clb.Root{ID: taxonID}
	}

	exportKey, err := client.TriggerExport(datasetID, er)
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

func (c *clbSource) Extract(_ string) error {
	return nil
}

func (c *clbSource) InitSfga() (sfga.Archive, error) {
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

func (c *clbSource) ToSfga(_ sfga.Archive) error {
	return nil
}

// resolveTaxonID determines the root taxon ID for the export.
// Priority: CLI flag > dataset default (resolved by name) > none.
func (c *clbSource) resolveTaxonID(
	client *clb.Client,
	datasetID int,
) (string, error) {
	if c.cfg.CLBTaxonID != "" {
		return c.cfg.CLBTaxonID, nil
	}

	if c.ds.TaxonName != "" {
		id, err := client.SearchTaxon(
			datasetID, c.ds.TaxonName,
			c.ds.TaxonRank, "accepted",
		)
		if err != nil {
			return "", fmt.Errorf("resolving root taxon: %w", err)
		}
		return id, nil
	}

	return "", nil
}
