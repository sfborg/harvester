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

// clbDataset defines a featured ChecklistBank dataset.
type clbDataset struct {
	// Label is the short identifier used in `harvester get <label>`.
	Label string

	// Name is the human-readable dataset name.
	Name string

	// Notes is the description shown in `harvester list -v`.
	Notes string

	// DatasetID is the ChecklistBank dataset key.
	DatasetID int

	// TaxonName is the default root taxon resolved by name search.
	// Empty means no default root taxon filter.
	TaxonName string

	// TaxonRank narrows the taxon name search (e.g. "class").
	TaxonRank string
}

// registry lists featured ChecklistBank datasets that appear as
// individual entries in `harvester list`. Add new entries here.
var registry = []clbDataset{
	{
		Label:     "wsc",
		Name:      "World Spider Catalog",
		DatasetID: 56185,
		TaxonName: "Arachnida",
		TaxonRank: "class",
		Notes: `World Spider Catalog from ChecklistBank (dataset 56185).
Data is downloaded automatically via the ChecklistBank API as a
ColDP export. Requires --clb-user and --clb-password flags for
authentication.

Root taxon 'Arachnida' is resolved dynamically by name search.
Use --clb-taxon-id to override with a specific taxon ID.`,
	},
	{
		Label: "scarabs",
		Name: "World Scarabaeidae Database",
		DatasetID: 1027,
		TaxonName: "Scarabaeoidea",
		TaxonRank: "superfamily",
		Notes: "",
	},
}

// IsCLBSource returns true if the given label belongs to the generic
// CLB source or any featured dataset in the registry.
func IsCLBSource(label string) bool {
	if label == "clb" {
		return true
	}
	for _, d := range registry {
		if d.Label == label {
			return true
		}
	}
	return false
}

// NewAll returns convertors for the generic CLB source plus every
// featured dataset in the registry.
func NewAll(cfg config.Config) []data.Convertor {
	res := []data.Convertor{newGeneric(cfg)}
	for _, d := range registry {
		res = append(res, newRegistered(cfg, d))
	}
	return res
}

// --- generic CLB source (requires --clb-dataset-id) ---

func newGeneric(cfg config.Config) data.Convertor {
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
	return &clbSource{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
}

// --- registered (featured) CLB source ---

func newRegistered(cfg config.Config, d clbDataset) data.Convertor {
	set := data.DataSet{
		Label:       d.Label,
		Name:        d.Name,
		Notes:       d.Notes,
		ManualSteps: false,
	}
	return &clbSource{
		cfg:       cfg,
		ds:        &d,
		Convertor: base.New(cfg, &set),
	}
}

// --- shared implementation ---

type clbSource struct {
	data.Convertor
	cfg       config.Config
	ds        *clbDataset // nil for the generic source
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

	datasetID := c.datasetID()
	if datasetID == 0 {
		return "", fmt.Errorf(
			"--clb-dataset-id is required for the clb source",
		)
	}

	if c.cfg.CLBUser == "" || c.cfg.CLBPassword == "" {
		return "", fmt.Errorf(
			"--clb-user and --clb-password are required",
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

// datasetID returns the dataset ID from the registry entry or
// from the CLI flag.
func (c *clbSource) datasetID() int {
	if c.ds != nil {
		return c.ds.DatasetID
	}
	return c.cfg.CLBDatasetID
}

// resolveTaxonID determines the root taxon ID for the export.
// Priority: CLI flag > registry default (resolved by name) > none.
func (c *clbSource) resolveTaxonID(
	client *clb.Client,
	datasetID int,
) (string, error) {
	// CLI flag takes priority.
	if c.cfg.CLBTaxonID != "" {
		return c.cfg.CLBTaxonID, nil
	}

	// Registered source with a default taxon: resolve by name.
	if c.ds != nil && c.ds.TaxonName != "" {
		id, err := client.SearchTaxon(
			datasetID, c.ds.TaxonName,
			c.ds.TaxonRank, "accepted",
		)
		if err != nil {
			return "", fmt.Errorf("resolving root taxon: %w", err)
		}
		return id, nil
	}

	// Generic source with no taxon specified.
	return "", nil
}
