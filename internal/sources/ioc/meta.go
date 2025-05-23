package ioc

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sfborg/sflib/pkg/coldp"
)

func (l *ioc) importMeta() error {
	meta, err := l.metaFromFile()
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
	if l.cfg.ArchiveDate != "" {
		meta.Issued = l.cfg.ArchiveDate
	}
	l.sfga.InsertMeta(meta)
	return nil
}

func (l *ioc) metaFromFile() (*coldp.Meta, error) {
	res := coldp.Meta{}
	f, err := os.Open(l.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	scanner.Scan()
	scanner.Scan()
	line := scanner.Text()
	scanner.Scan()

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

	// save the rest of the file to a tsv file without metadata
	path := filepath.Join(l.cfg.ExtractDir, "data.tsv")
	wFile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer wFile.Close()
	w := bufio.NewWriter(wFile)
	defer w.Flush()

	for scanner.Scan() {
		_, err = w.WriteString(scanner.Text() + "\n")
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	l.path = path

	return &res, nil
}
