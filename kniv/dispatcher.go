package kniv

import (
	"fmt"
	"log"
)

type registeredProcessor struct {
	Id            uint
	Name          string
	consumeLabels []Label
	produceLabels []Label
	processor     Processor
}

func (r *registeredProcessor) addConsumeLabels(consumeLabels []Label) {
	r.consumeLabels = append(r.consumeLabels, consumeLabels...)
}

func (r *registeredProcessor) addProduceLabels(produceLabels []Label) {
	r.produceLabels = append(r.produceLabels, produceLabels...)
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

func (rs registeredProcessors) get(name string) (*registeredProcessor, bool) {
	for _, r := range rs {
		if r.Name == name {
			return r, true
		}
	}
	return nil, false
}

func (rs registeredProcessors) getById(id uint) (*registeredProcessor, bool) {
	for _, r := range rs {
		if r.Id == id {
			return r, true
		}
	}
	return nil, false
}

func (rs registeredProcessors) addConsumeLabels(id uint, consumeLabels []Label) bool {
	processor, ok := rs.getById(id)
	if !ok {
		return false
	}
	processor.addConsumeLabels(consumeLabels)
	return true
}

func (rs registeredProcessors) addProduceLabels(id uint, produceLabels []Label) bool {
	processor, ok := rs.getById(id)
	if !ok {
		return false
	}
	processor.addProduceLabels(produceLabels)
	return true
}

type Dispatcher struct {
	registeredProcessors registeredProcessors
	queue                chan Event
	produceLabelMap      map[uint64][]Label
	eventId              uint64
	processorId          uint
}

func NewDispatcher(queueSize int) *Dispatcher {
	return &Dispatcher{
		queue:           make(chan Event, queueSize),
		produceLabelMap: map[uint64][]Label{},
		eventId:         0,
	}
}

func (d *Dispatcher) RegisterProcessor(name string, consumeLabels, produceLabels []Label, processor Processor) uint {
	processor.SetOutChannel(d.queue)
	d.processorId++
	d.registeredProcessors = append(d.registeredProcessors, &registeredProcessor{
		Id:            d.processorId,
		Name:          name,
		consumeLabels: consumeLabels,
		produceLabels: produceLabels,
		processor:     processor,
	})
	return d.processorId
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
		log.Printf("%d -> %d: new event: %#v\n", event.GetSourceId(), event.GetId(), event)

		consumedLabel := event.PopLabel()

		filteredProcessors := d.registeredProcessors.filterByConsumeLabel(consumedLabel)

		if len(filteredProcessors) == 0 {
			log.Println(consumedLabel + " not found")
			continue
		}

		for _, filteredProcessor := range filteredProcessors {
			//if i > 0 {
			newEvent := event.Copy() // FIXME Copy is not complete implement
			newEvent.PushRoute(fmt.Sprintf("%d->%d:%s", event.GetSourceId(), event.GetId(), consumedLabel))
			d.eventId++
			newEvent.SetId(d.eventId)
			newEvent.SetSourceId(event.GetId())
			//}
			log.Printf("%d -> %d: fork %#v\n", event.GetId(), newEvent.GetId(), newEvent)
			msg := fmt.Sprintf("%d -> %d: sent to %s: %s -> %s", event.GetId(), newEvent.GetId(), filteredProcessor.processor.GetName(), newEvent.GetLabels(), consumedLabel)

			produceLabels := filteredProcessor.produceLabels
			if len(produceLabels) > 0 {
				msg += fmt.Sprintf(" <- %s", produceLabels)
				newEvent.PushLabels(produceLabels)
			}
			log.Println(msg)
			filteredProcessor.processor.Enqueue(newEvent)
		}
	}
}

func (d *Dispatcher) StartProcessors() {
	d.registeredProcessors.start()
}

func (d *Dispatcher) GetProcessor(name string) (*registeredProcessor, bool) {
	return d.registeredProcessors.get(name)
}
