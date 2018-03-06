package kniv

import (
	"fmt"
	"log"
)

type BaseProcessor struct {
	Name    string
	inChan  chan Resource
	outChan chan Resource
	Process func(resource Resource) (Resource, error)
}

func (b *BaseProcessor) GetName() string {
	return b.Name
}

func (b *BaseProcessor) Enqueue(resource Resource) {
	b.inChan <- resource
}

func (b *BaseProcessor) SetOutChannel(outChan chan Resource) {
	b.outChan = outChan
}

func (b *BaseProcessor) Start() {
	for resource := range b.inChan {
		fmt.Println(b.GetName(), "start processing:", resource)
		processedResource, err := b.Process(resource)
		if err != nil {
			// TODO: Add err chan
			log.Println(err)
			continue
		}
		b.outChan <- processedResource
	}
}

type Processor interface {
	GetName() string
	Enqueue(resource Resource)
	SetOutChannel(outChan chan Resource)
	Start()
}
