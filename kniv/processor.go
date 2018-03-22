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

type BaseArgs struct {
	QueueSize int
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
	GetName() string
	Enqueue(resource Event)
	SetOutChannel(outChan chan Event)
	Start()
}
