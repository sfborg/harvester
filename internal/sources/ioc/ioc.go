package ioc

type ioc struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
	path string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "grin",
		Name:  "GRIN Plant Taxonomy",
		Notes: `Create tsv file from current master file at 
https://www.worldbirdnames.org/new/ioc-lists/master-list-2/
and save to the box.com, generate new URL and update it here.`,
		ManualSteps: true,
		URL:         "https://uofi.box.com/shared/static/x9f7o161l81my22by0k8ov2kgfmuuunu.tsv",
	}
	res := ioc{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (i *ioc) Extract(path string) error {
	slog.Info("Importing IOC world birds list to a temporary SQLite database")
	file := filepath.Base(path)
	g.path = filepath.Join(g.cfg.ExtractDir, file)
	_, err := gnsys.CopyFile(path, g.path)
	if err != nil {
		return err
	}
	return nil
}
