package kniv

import (
	"log"
)

type BaseProcessor struct {
	Name    string
	inChan  chan Event
	outChan chan Event
	Process func(resource Event) ([]Event, error)
}

func (b *BaseProcessor) GetName() string {
	return b.Name
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
		processedEvent, err := b.Process(event)
		if err != nil {
			// TODO: Add err chan
			log.Println(err)
			continue
		}
		for _, e := range processedEvent {
			e.SetSourceId(sourceEventId)
			e.SetLabels(event.GetLabels())
			b.outChan <- e
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
	GetName() string
	Enqueue(resource Event)
	SetOutChannel(outChan chan Event)
	Start()
}
