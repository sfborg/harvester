package paleodb

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sfborg/harvester/internal/sysio"
)

func (p *paleodb) Download() (string, error) {
	var err error
	err = sysio.ResetCache(p.cfg)
	if err != nil {
		return "", err
	}
	taxaURL := p.set.URL + "/taxa/list.txt?all_taxa=true&show=attr,app,common,parent,immparent,classext,ecospace,ttaph,img,ref,refattr,ent,entname,crmod"
	resp, err := http.Get(taxaURL)
	if err != nil {
		return "", err
	}

	dir := p.cfg.ExtractDir
	taxon := filepath.Join(dir, "taxon.json")
	if err != nil {
		return "", err
	}

	var bs []byte
	_, err = resp.Body.Read(bs)
	fmt.Println("body")
	fmt.Println(string(bs))

	if err != nil {
		return "", err
	}

	err = os.WriteFile(taxon, bs, 0644)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return "", nil
}

func (p *paleodb) Import(_ string) error {
	return nil
}
