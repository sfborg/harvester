package data

type Set struct {
	Label       string
	Name        string
	Description string
	ManualSteps bool
	URL         string
	New         func(Set) Convertor
}

type Convertor interface {
	Label() string
	Name() string
	Description() string
	ManualSteps() bool
	Download() error
	ToSFGA() error
}
