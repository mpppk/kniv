package kniv

import (
	"fmt"
	"log"
)

type registeredProcessor struct {
	resourceType ResourceType
	processor    Processor
}

type registeredProcessors []*registeredProcessor

func (rs registeredProcessors) add(resourceType ResourceType, processor Processor) {
	rs = append(rs, &registeredProcessor{
		resourceType: resourceType,
		processor:    processor,
	})
}

func (rs registeredProcessors) filter(resourceType ResourceType) registeredProcessors {
	var ret registeredProcessors
	for _, r := range rs {
		if r.resourceType == resourceType {
			ret = append(ret, r)
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
	queue                chan Resource
}

func NewDispatcher(queueSize int) *Dispatcher {
	return &Dispatcher{
		queue: make(chan Resource, queueSize),
	}
}

func (d *Dispatcher) RegisterProcessor(resourceType ResourceType, processor Processor) {
	processor.SetOutChannel(d.queue)
	d.registeredProcessors = append(d.registeredProcessors, &registeredProcessor{
		resourceType: resourceType,
		processor:    processor,
	})
}

func (d *Dispatcher) AddResource(resource Resource) {
	d.queue <- resource
}

func (d *Dispatcher) Start() {
	for resource := range d.queue {
		log.Println("new resource:", resource)
		filteredProcessors := d.registeredProcessors.filter(resource.ResourceType)
		if len(filteredProcessors) == 0 {
			log.Println(resource.ResourceType + " not found")
			continue
		}

		for _, processor := range filteredProcessors.toProcessors() {
			fmt.Println("resource consumed by ", processor.GetName())
			fmt.Println(&processor)
			processor.Enqueue(resource)
		}
	}
}

func (d *Dispatcher) StartProcessors() {
	d.registeredProcessors.start()
}
