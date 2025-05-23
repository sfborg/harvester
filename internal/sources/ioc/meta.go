package ioc

import (
	"bufio"
	"os"
	"regexp"

	"github.com/gnames/coldp/ent/coldp"
)

func (i *ioc) importMeta() error {
	meta, err := i.metaFromFile()
	if err != nil {
		return err
	}
	meta.Title = "IOC World Bird List"
	meta.Description = "The IOC World Bird List is an open access resource of " +
		"the international community of ornithologists. Our primary goal " +
		"is to facilitate worldwide communication in ornithology and " +
		"conservation based on an up-to-date evolutionary classification " +
		"of world birds and a set of English names that follow explicit " +
		"guidelines for spelling and construction."

	meta.URL = "https://www.worldbirdnames.org"
	if i.cfg.ArchiveDate != "" {
		meta.Issued = i.cfg.ArchiveDate
	}
	i.sfga.InsertMeta(meta)
	return nil
}

func (i *ioc) metaFromFile() (*coldp.Meta, error) {
	res := coldp.Meta{}
	f, err := os.Open(i.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	linesRead := 0

	scanner.Scan()
	scanner.Scan()
	line := scanner.Text()

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`\s+(.*)IOC World Bird List\s\(([^)]+)\)\. Doi\s(.*)\.\s`)
	match := re.FindStringSubmatch(line)

	if len(match) > 0 {
		res.Citation = match[1]
		res.Version = match[2]
		res.DOI = match[3]
	}

	return &res, nil
}
