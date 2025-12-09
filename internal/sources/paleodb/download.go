package paleodb

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gnames/gn"
	"github.com/sfborg/harvester/internal/sysio"
)

func (p *paleodb) Download() (string, error) {
	var err error
	if p.cfg.SkipDownload {
		return "", nil
	}

	err = sysio.ResetCache(p.cfg)
	if err != nil {
		return "", err
	}
	ctx := context.Background()

	slog.Info("readilng taxonomy data")
	gn.Info("Readilng taxonomy data")
	taxaURL := p.set.URL + "/taxa/list.txt?all_taxa=true&show=attr,app,common,parent,immparent,classext,ecospace,ttaph,img,ref,refattr,ent,entname,crmod"
	taxonFile := filepath.Join(p.cfg.ExtractDir, "taxon.csv")
	err = p.httpRequest(ctx, taxaURL, taxonFile)
	if err != nil {
		return "", err
	}

	slog.Info("readilng specimen data")
	gn.Info("Readilng specimen data")
	specURL := p.set.URL + "/specs/list.txt?all_records=true&show=attr,abund,plant,ecospace,taphonomy,coll,coords,loc,strat,lith,methods,env,geo,rem,resgroup,ent,entname,crmod"
	specFile := filepath.Join(p.cfg.ExtractDir, "spec.csv")
	err = p.httpRequest(ctx, specURL, specFile)
	if err != nil {
		return "", err
	}

	slog.Info("readilng reference data")
	gn.Info("Readilng reference data")
	refURL := p.set.URL + "/refs/list.json?vocab=bibjson&all_records=true"
	refFile := filepath.Join(p.cfg.ExtractDir, "ref.json")
	err = p.httpRequest(ctx, refURL, refFile)
	if err != nil {
		return "", err
	}

	slog.Info("readilng ranks")
	gn.Info("Readilng ranks")
	ranksURL := p.set.URL + "/config.txt?show=ranks"
	ranksFile := filepath.Join(p.cfg.ExtractDir, "ranks.csv")
	err = p.httpRequest(ctx, ranksURL, ranksFile)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (p *paleodb) Extract(_ string) error {
	return nil
}

func (p *paleodb) httpRequest(ctx context.Context, url, file string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/plain")

	resp, err := p.http.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var respBytes []byte
	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(file, respBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
