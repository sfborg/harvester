package harvester

type Harvester interface {
	List() []string
	Convert(dataset string) error
}
