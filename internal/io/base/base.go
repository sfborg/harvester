package base

import "github.com/sfborg/harvester/internal/ent/data"

// Convertor implements default methods of data.Convertor interface.
type Convertor struct {
	set data.Set
}

func New(s data.Set) data.Convertor {
	res := Convertor{set: s}
	return &res
}

func (c *Convertor) Label() string {
	return c.set.Label
}

func (c *Convertor) Name() string {
	return c.set.Name
}

func (c *Convertor) Description() string {
	return c.set.Description
}
func (c *Convertor) ManualSteps() bool {
	return c.set.ManualSteps
}

func (c *Convertor) Download() error {
	return nil
}

func (c *Convertor) ToSFGA() error {
	return nil
}
