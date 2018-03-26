package twitter

import (
	"github.com/mitchellh/mapstructure"
	"github.com/mpppk/kniv/kniv"
)

type ProcessorGenerator struct{}

func (g *ProcessorGenerator) Generate(intfArgs interface{}) (kniv.Processor, error) {
	var config Config
	err := mapstructure.Decode(intfArgs, &config)
	if err != nil {
		return nil, err
	}
	return NewProcessor(100000, &config), nil // FIXME queue size
}

func (g *ProcessorGenerator) GetType() string {
	return processorType
}
