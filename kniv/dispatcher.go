package kniv

import (
	"fmt"
	"log"
)

type Dispatcher struct {
	processorMap map[ResourceType]Processor
	queue        chan Resource
}

func NewDispatcher(queueSize int) *Dispatcher {
	return &Dispatcher{
		processorMap: map[ResourceType]Processor{},
		queue:        make(chan Resource, queueSize),
	}
}

func (d *Dispatcher) RegisterProcessor(resourceType ResourceType, processor Processor) {
	processor.SetOutChannel(d.queue)
	d.processorMap[resourceType] = processor

}

func (d *Dispatcher) AddResource(resource Resource) {
	d.queue <- resource
}

func (d *Dispatcher) Start() {
	for resource := range d.queue {
		log.Println("new resource:", resource)
		processor, ok := d.processorMap[resource.ResourceType]
		if !ok {
			log.Println(resource.ResourceType + " not found")
			continue
		}
		fmt.Println("resource consumed by ", processor.GetName())
		fmt.Println(&processor)
		processor.Enqueue(resource)
	}
}

func (d *Dispatcher) StartProcessors() {
	for _, p := range d.processorMap {
		go p.Start()
	}
}
