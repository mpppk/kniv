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
	ProcessorName string `yaml:"processor"`
	Name          string
	Args          interface{}
}

type Pipeline struct {
	Name string
	Jobs []*Job
}

type Job struct {
	Processor string
	Consume   []Label
	Produce   []Label
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
