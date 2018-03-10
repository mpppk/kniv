package kniv

import (
	"log"
)

type registeredProcessor struct {
	consumeLabels []Label
	produceLabels []Label
	processor     Processor
}

type registeredProcessors []*registeredProcessor

func (rs registeredProcessors) add(consumeLabels, produceLabels []Label, processor Processor) {
	rs = append(rs, &registeredProcessor{
		consumeLabels: consumeLabels,
		produceLabels: produceLabels,
		processor:     processor,
	})
}

func (rs registeredProcessors) filterByConsumeLabel(label Label) registeredProcessors {
	var ret registeredProcessors
	for _, r := range rs {
		for _, consumeLabel := range r.consumeLabels {
			if consumeLabel == label {
				ret = append(ret, r)
			}
		}
	}
	return ret
}

func (rs registeredProcessors) toProcessors() (processors []Processor) {
	for _, r := range rs {
		processors = append(processors, r.processor)
	}
	return processors
}

func (rs registeredProcessors) start() {
	for _, p := range rs.toProcessors() {
		go p.Start()
	}
}

type Dispatcher struct {
	registeredProcessors registeredProcessors
	queue                chan Event
}

func NewDispatcher(queueSize int) *Dispatcher {
	return &Dispatcher{
		queue: make(chan Event, queueSize),
	}
}

func (d *Dispatcher) RegisterProcessor(consumeLabels, produceLabels []Label, processor Processor) {
	processor.SetOutChannel(d.queue)
	d.registeredProcessors = append(d.registeredProcessors, &registeredProcessor{
		consumeLabels: consumeLabels,
		produceLabels: produceLabels,
		processor:     processor,
	})
}

func (d *Dispatcher) AddResource(event Event) {
	d.queue <- event
}

func (d *Dispatcher) Start() {
	for event := range d.queue {
		log.Printf("new event: %#v", event)
		filteredProcessors := d.registeredProcessors.filterByConsumeLabel(event.GetLatestLabel())
		if len(filteredProcessors) == 0 {
			log.Println(event.GetLatestLabel() + " not found")
			continue
		}

		for _, processor := range filteredProcessors.toProcessors() {
			log.Println("event is sent to", processor.GetName())
			processor.Enqueue(event)
		}
	}
}

func (d *Dispatcher) StartProcessors() {
	d.registeredProcessors.start()
}
