package ott

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"
)

func (o *ott) collectSynonyms() error {
	var g errgroup.Group
	chIn := make(chan []string)

	g.Go(func() error {
		return o.processSynonyms(chIn)
	})

	g.Go(func() error {
		defer close(chIn)
		return o.loadSynonyms(chIn)
	})

	return g.Wait()
}

func (o *ott) loadSynonyms(chIn chan<- []string) error {
	f, err := os.Open(o.synonymsPath)
	if err != nil {
		return fmt.Errorf("open synonyms: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		line = strings.TrimRight(line, "\t| ")
		fields := strings.Split(line, "\t|\t")
		if len(fields) < 2 {
			continue
		}
		chIn <- fields
	}
	return scanner.Err()
}

func (o *ott) processSynonyms(chIn <-chan []string) error {
	for row := range chIn {
		name := strings.TrimSpace(row[0])
		uid := strings.TrimSpace(row[1])
		if name == "" || uid == "" {
			continue
		}
		synType := ""
		if len(row) > 2 {
			synType = strings.TrimSpace(row[2])
		}
		o.synonyms[uid] = append(o.synonyms[uid], synDatum{
			name:    name,
			synType: synType,
		})
	}
	return nil
}
