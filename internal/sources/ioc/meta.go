package ioc

import (
	"bufio"
	"fmt"
	"log"
	"os"

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
	f, err := os.Open(i.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	linesRead := 0

	for scanner.Scan() { // Scan returns false when the end of the file is reached or an error occurs
		fmt.Printf("Read line %d: %s\n", linesRead+1, scanner.Text())
		linesRead++

		if linesRead >= 3 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}
