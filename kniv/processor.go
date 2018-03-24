package kniv

import (
	"fmt"
	"log"
)

type BaseProcessor struct {
	Name    string
	Type    string
	inChan  chan Event
	outChan chan Event
	Process func(resource Event) ([]Event, error)
}

type BaseArgs struct {
	QueueSize int
}

func (b *BaseProcessor) GetName() string {
	return b.Name
}

func (b *BaseProcessor) SetName(name string) {
	b.Name = name
}

func (b *BaseProcessor) GetType() string {
	return b.Type
}

func (b *BaseProcessor) Enqueue(resource Event) {
	b.inChan <- resource
}

func (b *BaseProcessor) SetOutChannel(outChan chan Event) {
	b.outChan = outChan
}

func (b *BaseProcessor) Start() {
	for event := range b.inChan {
		log.Printf("%s has been started event processing: %#v", b.GetName(), event)
		sourceEventId := event.GetId()
		processedEvents, err := b.Process(event)
		if err != nil {
			// TODO: Add err chan
			log.Println(err)
			continue
		}
		if processedEvents == nil {
			log.Printf("%d: filtered", event.GetId())
			return
		}

		for _, e := range processedEvents {
			if e != nil {
				e.SetSourceId(sourceEventId)
				e.SetLabels(event.CopyLabels())
				e.SetRoutes(event.CopyRoutes())
				b.outChan <- e
			} else {
				log.Printf("%d: filtered!!", event.GetId())
			}
		}
	}
}

func NewBaseProcessor(queueSize int) *BaseProcessor {
	return &BaseProcessor{
		Name:   "base processor",
		inChan: make(chan Event, queueSize),
	}
}

type Processor interface {
	GetType() string
	GetName() string
	SetName(name string)
	Enqueue(resource Event)
	SetOutChannel(outChan chan Event)
	Start()
}

type ProcessorGenerator interface {
	GetType() string
	Generate(intfArgs interface{}) (Processor, error)
}

type ProcessorFactory struct {
	generators []ProcessorGenerator
}

func (pf *ProcessorFactory) AddGenerator(generator ProcessorGenerator) {
	pf.generators = append(pf.generators, generator)
}

func (pf *ProcessorFactory) Create(setting FlowSetting) (Processor, error) {
	for _, generator := range pf.generators {
		if setting.GetProcessorType() == generator.GetType() {
			processor, err := generator.Generate(setting.GetArgs())
			if err != nil {
				return nil, err
			}
			processor.SetName(setting.GetName())
			return processor, nil
		}
	}
	return nil, fmt.Errorf("processor type %s not found", setting.GetProcessorType())
}
