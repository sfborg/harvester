package ncbi

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"
)

func (n *ncbi) collectNames() error {
	var g errgroup.Group
	var err error
	chIn := make(chan []string)

	g.Go(func() error {
		return n.processNames(chIn)
	})

	g.Go(func() error {
		defer close(chIn)
		err = n.loadNames(chIn)
		return err
	})

	if err = g.Wait(); err != nil {
		return err
	}
	return nil
}

func (n *ncbi) loadNames(chIn chan<- []string) error {
	file, err := os.Open(n.namePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSuffix(line, "\t|")
		fields := strings.Split(line, "\t|\t")
		if len(fields) != 4 {
			return fmt.Errorf("wrong number of name fields: %s", line)
		}
		chIn <- fields
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (n *ncbi) processNames(chIn <-chan []string) error {
	for row := range chIn {
		id := row[0]
		name := row[1]
		nameType := row[3]
		if nameType == "scientific name" {
			nameType = "valid"
		}

		if _, ok := n.names[id]; !ok {
			n.names[id] = make(map[string]string)
		}
		n.names[id][nameType] = name
	}
	// available types in 2025
	// acronym
	// authority
	// blast name
	// common name
	// equivalent name
	// genbank common name
	// in-part
	// includes
	// synonym
	// type material
	// valid
	return nil
}
