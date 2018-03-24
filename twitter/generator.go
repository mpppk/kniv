package twitter

import (
	"fmt"
	"github.com/mpppk/kniv/kniv"
)

type ProcessorGenerator struct{}

func (g *ProcessorGenerator) Generate(intfArgs interface{}) (kniv.Processor, error) {
	config, ok := intfArgs.(Config)
	if !ok {
		return nil, fmt.Errorf("invalid delay processor config: %#v", intfArgs)
	}
	return NewProcessor(100000, &config), nil // FIXME queue size
}

func (g *ProcessorGenerator) GetType() string {
	return processorType
}
