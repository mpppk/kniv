package twitter

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mpppk/kniv/kniv"
	"os"
	"strings"
)

type ProcessorGenerator struct{}

func (g *ProcessorGenerator) Generate(intfArgs interface{}) (kniv.Processor, error) {
	mapArgs, ok := intfArgs.(map[interface{}]interface{})

	// FIXME convert to util method
	// Set environment if value start with $
	for k, v := range mapArgs {
		if vStr, vOk := v.(string); vOk {
			if strings.HasPrefix(vStr, "$") {
				fmt.Println(vStr[1:])
				mapArgs[k] = os.Getenv(vStr[1:])
			}
		}
	}

	if !ok {
		return nil, errors.New("invalid args are passed to twitter processor")
	}

	var config Config
	err := mapstructure.Decode(mapArgs, &config)
	if err != nil {
		return nil, err
	}
	return NewProcessor(100000, &config), nil // FIXME queue size
}

func (g *ProcessorGenerator) GetType() string {
	return processorType
}
