package kniv

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Flow struct {
	Pipelines  []*Pipeline
	Processors []*ProcessorSetting
}

type FlowSetting interface {
	GetProcessorType() string
	GetName() string
	GetArgs() interface{}
}

type ProcessorSetting struct {
	ProcessorType string `yaml:"processor"`
	Name          string
	Args          interface{}
}

func (p *ProcessorSetting) GetProcessorType() string {
	return p.ProcessorType
}

func (p *ProcessorSetting) GetName() string {
	return p.Name
}

func (p *ProcessorSetting) GetArgs() interface{} {
	return p.Args
}

type Pipeline struct {
	Name string
	Jobs []*Job
}

type Job struct {
	ProcessorSetting `yaml:",inline"`
	Consume          []Label
	Produce          []Label
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
