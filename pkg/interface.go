package harvester

type Harvester interface {
	List() []string
	Get(datasetLabel, outPath string) error
}
