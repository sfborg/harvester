package harvester

type Harvester interface {
	List() []string
	Convert(dataset, path string) error
}
