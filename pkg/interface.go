package harvester

type Harvester interface {
	List() []string
	Convert(datasetLabel, outPath string) error
}
