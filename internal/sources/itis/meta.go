package itis

import (
	"github.com/sfborg/sflib/pkg/coldp"
)

func (t *itis) importMeta() error {
	meta := &coldp.Meta{
		Title: "Integrated Taxonomic Information System (ITIS)",
		Description: "ITIS is a partnership of federal agencies and other " +
			"organizations designed to provide scientifically credible " +
			"taxonomic information. The ITIS database contains taxonomic " +
			"information on plants, animals, fungi, and microbes of " +
			"North America and the world.",
		URL: "https://www.itis.gov",
	}

	if t.cfg.ArchiveDate != "" {
		meta.Issued = t.cfg.ArchiveDate
	}

	// Try to get version from database metadata if available.
	version, err := t.getVersion()
	if err == nil && version != "" {
		meta.Version = version
	}

	t.sfga.InsertMeta(meta)
	return nil
}

func (t *itis) getVersion() (string, error) {
	q := `SELECT version FROM version LIMIT 1`
	var version string
	err := t.db.QueryRow(q).Scan(&version)
	if err != nil {
		return "", err
	}
	return version, nil
}
