package ncbi

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"
)

func (n *ncbi) collectNodes() error {
	var g errgroup.Group
	var err error
	chIn := make(chan []string)

	g.Go(func() error {
		return n.processNodes(chIn)
	})

	g.Go(func() error {
		defer close(chIn)
		err = n.loadNodes(chIn)
		return err
	})

	if err = g.Wait(); err != nil {
		return err
	}
	return nil
}

func (n *ncbi) loadNodes(chIn chan<- []string) error {
	file, err := os.Open(n.nodePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSuffix(line, "\t|")
		fields := strings.Split(line, "\t|\t")
		if len(fields) != 13 {
			return fmt.Errorf("wrong number of nodes fields: %d, %s",
				len(fields), line,
			)
		}
		chIn <- fields
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (n *ncbi) processNodes(chIn <-chan []string) error {
	for row := range chIn {
		id := row[0]
		if id == "1" {
			continue
		}

		parentID := row[1]
		rank := row[2]
		if rank == "no rank" {
			rank = ""
		}
		if parentID == "1" {
			parentID = id
		}
		if _, ok := n.names[id]; !ok {
			continue
		}
		if _, ok := n.names[id]["valid"]; !ok {
			continue
		}

		var vernNames []string
		var synonyms []synonym
		for k := range n.names[id] {
			if slices.Contains(nameType, k) {
				continue
			}
			name := n.names[id][k]
			if slices.Contains(vernType, k) {
				vernNames = append(vernNames, name)
			} else if k == "synonym" {
				synonyms = append(synonyms, synonym{name: name, taxonomicStatus: k})
			}
		}
		nameString := n.names[id]["valid"]
		canonical := nameString
		if au, ok := n.names[id]["authority"]; ok {
			nameString = au
		}
		rec := datum{
			taxonID:   id,
			parentID:  parentID,
			canonical: canonical,
			nameStr:   nameString,
			rank:      rank,
			vernNames: vernNames,
			synonyms:  synonyms,
		}
		n.data = append(n.data, rec)

	}
	return nil
}
