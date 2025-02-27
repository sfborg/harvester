package base

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/sysio"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/ent/sfga"
	"github.com/sfborg/sflib/io/sfgaio"
)

// Convertor implements default methods of data.Convertor interface.
type Convertor struct {
	set *data.Set
	cfg config.Config
	gnp gnparser.GNparser
}

func New(cfg config.Config, s *data.Set) data.Convertor {
	res := Convertor{
		cfg: cfg,
		set: s,
	}
	gncfg := gnparser.NewConfig(
		gnparser.OptCode(cfg.Code),
		gnparser.OptWithDetails(true),
	)
	res.gnp = gnparser.New(gncfg)
	return &res
}

func (c *Convertor) Config() config.Config {
	return c.cfg
}

func (c *Convertor) Label() string {
	return c.set.Label
}

func (c *Convertor) Name() string {
	return c.set.Name
}

func (c *Convertor) Description() string {
	return c.set.Description
}
func (c *Convertor) ManualSteps() bool {
	return c.set.ManualSteps
}

func (c *Convertor) Download() (string, error) {
	var err error
	var path string
	err = sysio.ResetCache(c.cfg)
	if err != nil {
		return "", err
	}

	if c.cfg.LoadFile != "" {
		slog.Info(
			"Using local file", "source", c.set.Label, "file", c.cfg.LoadFile,
		)
		return c.cfg.LoadFile, nil
	}

	if c.set.URL == "" {
		err = errors.New("no local file or URL given")
		return "", err
	}

	slog.Info("Downloading", "source", c.set.Label)
	path, err = gnsys.Download(c.set.URL, c.cfg.DownloadDir, true)
	if err != nil {
		return path, err
	}
	return path, nil
}

func (c *Convertor) Import(path string) error {
	var f gnsys.Extractor
	switch gnsys.GetFileType(path) {
	case gnsys.ZipFT:
		f = gnsys.ExtractZip
	case gnsys.TarFT:
		f = gnsys.ExtractTar
	case gnsys.TarGzFT:
		f = gnsys.ExtractTarGz
	case gnsys.TarBzFT:
		f = gnsys.ExtractTarBz2
	case gnsys.TarXzFt:
		f = gnsys.ExtractTarXz
	default:
		return fmt.Errorf("cannot determine file format of '%s'", path)
	}
	err := f(path, c.cfg.ExtractDir)
	if err != nil {
		return err
	}
	return nil
}

func (c *Convertor) InitSFGA() (sfga.Archive, error) {
	sysio.EmptyDir(c.cfg.SfgaDir)

	sfga := sfgaio.New()
	err := sfga.Create(c.cfg.SfgaDir)
	if err != nil {
		return nil, err
	}
	_, err = sfga.Connect()
	if err != nil {
		return nil, err
	}
	return sfga, nil
}

func (c *Convertor) Parse(name string) data.Parsed {
	parsedName := c.gnp.ParseName(name)

	result := data.Parsed{NameID: parsedName.VerbatimID}
	if !parsedName.Parsed {
		return result
	}

	result = data.Parsed{
		NameID:          parsedName.VerbatimID,
		Quality:         parsedName.ParseQuality,
		CanonicalFull:   parsedName.Canonical.Full,
		CanonicalSimple: parsedName.Canonical.Simple,
	}

	if parsedName.Authorship != nil {
		result.Authorship = parsedName.Authorship.Verbatim
	}

	if parsedName.Authorship != nil && parsedName.Authorship.Combination != nil {
		result.CombinationAuthorship = formatAuthors(parsedName.Authorship.Combination.Authors)
	}

	switch detail := parsedName.Details.(type) {
	case parsed.DetailsUninomial:
		result.Uninomial = detail.Uninomial.Value
	case parsed.DetailsSpecies:
		result.Genus = detail.Species.Genus
		result.Species = detail.Species.Species
	case parsed.DetailsInfraspecies:
		if len(detail.Infraspecies.Infraspecies) == 1 {
			result.Genus = detail.Infraspecies.Genus
			result.Species = detail.Infraspecies.Species.Species
			result.Rank = detail.Infraspecies.Infraspecies[0].Rank
			result.Infraspecies = detail.Infraspecies.Infraspecies[0].Value
		}
	}
	return result
}

func (c *Convertor) ToSFGA(_ sfga.Archive) error {
	slog.Info("Running a placeholder ToSFGA method")
	return nil
}

func formatAuthors(authorship []string) string {
	result := ""
	switch len(authorship) {

	case 1:
		result = authorship[0]
	case 2:
		result = strings.Join(authorship, " & ")
	default:
		result = strings.Join(authorship[0:len(authorship)-1], ", ")
		result = result + " & " + authorship[len(authorship)-1]
	}
	return result
}
