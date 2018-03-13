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
	produceLabelMap      map[uint64][]Label
	eventId              uint64
}

func NewDispatcher(queueSize int) *Dispatcher {
	return &Dispatcher{
		queue:           make(chan Event, queueSize),
		produceLabelMap: map[uint64][]Label{},
		eventId:         0,
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
		if len(event.GetLabels()) == 0 || event.GetLatestLabel() == "done" {
			log.Printf("source %d: event done: %#v", event.GetSourceId(), event)
			continue
		}

		d.eventId++
		event.SetId(d.eventId)
		log.Printf("%d -> %d: new event: %#v", event.GetSourceId(), event.GetId(), event)

		consumedLabel := event.PopLabel()
		log.Printf("%d: consume label: %s -> %s", event.GetId(), event.GetLabels(), consumedLabel)

		filteredProcessors := d.registeredProcessors.filterByConsumeLabel(consumedLabel)

		if len(filteredProcessors) == 0 {
			log.Println(consumedLabel + " not found")
			continue
		}

		for i, filteredProcessor := range filteredProcessors {
			log.Println("event is sent to", filteredProcessor.processor.GetName())
			if i > 0 {
				event = event.Copy() // FIXME Copy is not complete implement
				d.eventId++
				event.SetId(d.eventId)
			}

			produceLabels := filteredProcessor.produceLabels
			if len(produceLabels) > 0 {
				log.Printf("%d: produce labels: %s <- %s", event.GetId(), event.GetLabels(), produceLabels)
				event.PushLabels(produceLabels)
			}
			filteredProcessor.processor.Enqueue(event)
		}
	}
}

func (d *Dispatcher) StartProcessors() {
	d.registeredProcessors.start()
}
