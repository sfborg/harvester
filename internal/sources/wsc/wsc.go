package wsc

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

const (
	datasetID       = 56185
	defaultTaxonName = "Arachnida"
	defaultTaxonRank = "class"
)

type wsc struct {
	data.Convertor
	cfg       config.Config
	coldpPath string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "wsc",
		Name:  "World Spider Catalog",
		Notes: `World Spider Catalog from ChecklistBank (dataset 56185).
Data is downloaded automatically via the ChecklistBank API as a
ColDP export. Requires --clb-user and --clb-password flags for
authentication.

Root taxon 'Arachnida' is resolved dynamically by name search.
Use --clb-taxon-id to override with a specific taxon ID.`,
		ManualSteps: false,
	}
	res := wsc{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (w *wsc) Download() (string, error) {
	if w.cfg.SkipDownload {
		return "", nil
	}

	if w.cfg.LoadFile != "" {
		slog.Info("using local file", "file", w.cfg.LoadFile)
		gn.Info("Using local file: %s", w.cfg.LoadFile)
		w.coldpPath = w.cfg.LoadFile
		return w.cfg.LoadFile, nil
	}

	if w.cfg.CLBUser == "" || w.cfg.CLBPassword == "" {
		return "", fmt.Errorf(
			"--clb-user and --clb-password are required for WSC",
		)
	}

	err := sysio.ResetCache(w.cfg)
	if err != nil {
		return "", err
	}

	client := clb.New(w.cfg.CLBApi, w.cfg.CLBUser, w.cfg.CLBPassword)

	err = client.Login()
	if err != nil {
		return "", err
	}

	// Resolve root taxon ID. If --clb-taxon-id is set, use it
	// directly. Otherwise look up "Arachnida" by name so the export
	// keeps working even if the ID changes.
	taxonID := w.cfg.CLBTaxonID
	if taxonID == "" {
		taxonID, err = client.SearchTaxon(
			datasetID, defaultTaxonName,
			defaultTaxonRank, "accepted",
		)
		if err != nil {
			return "", fmt.Errorf("resolving root taxon: %w", err)
		}
	}

	er := clb.ExportRequest{
		Format:         w.cfg.CLBFormat,
		Synonyms:       w.cfg.CLBSynonyms,
		BareNames:      w.cfg.CLBBareNames,
		Extended:       w.cfg.CLBExtended,
		Extinct:        w.cfg.CLBExtinct,
		Classification: w.cfg.CLBClassification,
		TaxGroups:      w.cfg.CLBTaxGroups,
		MinRank:        w.cfg.CLBMinRank,
		TabFormat:      w.cfg.CLBTabFormat,
		Root:           &clb.Root{ID: taxonID},
	}

	exportKey, err := client.TriggerExport(datasetID, er)
	if err != nil {
		return "", err
	}

	downloadURL, err := client.WaitForExport(exportKey)
	if err != nil {
		return "", err
	}

	path, err := client.DownloadExport(downloadURL, w.cfg.DownloadDir)
	if err != nil {
		return "", err
	}

	w.coldpPath = path
	return path, nil
}

func (w *wsc) Extract(_ string) error {
	return nil
}

func (w *wsc) InitSfga() (sfga.Archive, error) {
	sysio.EmptyDir(w.cfg.SfgaDir)

	coldpPath := w.coldpPath
	if coldpPath == "" {
		var err error
		coldpPath, err = clb.FindColdpZip(w.cfg.DownloadDir)
		if err != nil {
			return nil, err
		}
	}

	return clb.ColdpToSfga(coldpPath, w.cfg)
}

func (w *wsc) ToSfga(_ sfga.Archive) error {
	return nil
}
