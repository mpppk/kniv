package kniv

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Flow struct {
	Pipelines  []*Pipeline
	Processors []*ProcessorSetting
}

type ProcessorSetting struct {
	Name string
	Args interface{}
}

type Pipeline struct {
	Name string
	Jobs []*Job
}

type Job struct {
	Processor string
	Consume   []string
	Produce   []string
	Args      interface{}
}

func LoadFlowFromFile(filepath string) *Flow {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	var flow Flow
	err = yaml.Unmarshal(buf, &flow)
	if err != nil {
		panic(err)
	}
	return &flow
}
