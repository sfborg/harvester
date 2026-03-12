package nzor

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type nzor struct {
	data.Convertor
	cfg       config.Config
	sfga      sfga.Archive
	http      *http.Client
	jsonlPath string
	donePath  string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "nzor",
		Name:  "New Zealand Organism Register",
		Notes: `NZOR is an actively maintained compilation of all organism names
relevant to New Zealand: indigenous, endemic or exotic species or species not
present in New Zealand but of national interest. Data is downloaded
automatically from the NZOR API.

Download is resumable: if interrupted, re-running will continue from the last
successfully downloaded page.`,
		ManualSteps: false,
		URL:         "https://data.nzor.org.nz/v1/names",
	}
	res := nzor{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
		http:      httpClient(),
		jsonlPath: filepath.Join(cfg.ExtractDir, "nzor.jsonl"),
		donePath:  filepath.Join(cfg.ExtractDir, "nzor.jsonl.done"),
	}
	return &res
}

func httpClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:      10,
		IdleConnTimeout:   600 * time.Second,
		ForceAttemptHTTP2: false,
	}
	return &http.Client{Timeout: 5 * time.Minute, Transport: tr}
}
