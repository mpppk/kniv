package kniv

import (
	"errors"
	"fmt"
)

type Label string

type EventPayload map[string]interface{}

func (e EventPayload) GetString(key string) (string, error) {
	value, ok := e[key]
	if !ok {
		return "", errors.New(key + " not found in event payload")
	}

	strValue, ok := value.(string)
	if !ok {
		return "", errors.New(key + " is not string in event payload")
	}
	return strValue, nil
}

func (e EventPayload) HasKey(key string) bool {
	_, exist := e[key]
	return exist
}

func (e EventPayload) HasEveryPayloadKeys(keys []string) bool {
	for _, key := range keys {
		if !e.HasKey(key) {
			return false
		}
	}
	return true
}

func (e EventPayload) HasSomePayloadKeys(keys []string) bool {
	for _, key := range keys {
		if e.HasKey(key) {
			return true
		}
	}
	return false
}

type Event interface {
	GetId() uint64
	SetId(uint64)
	GetSourceId() uint64
	SetSourceId(uint64)
	PushRoute(route string)
	GetRoutes() []string
	SetRoutes([]string)
	CopyRoutes() []string
	PopLabel() Label
	PushLabel(label Label)
	PushLabels(labels []Label)
	SetLabels(labels []Label)
	CopyLabels() []Label
	GetProduceLabels() []Label
	PushProduceLabels()
	GetLatestLabel() Label
	GetLabels() []Label
	SetPayload(payload EventPayload)
	GetPayload() EventPayload
	Copy() Event
}

type BaseEvent struct {
	id            uint64
	sourceId      uint64
	routes        []string
	labels        []Label
	produceLabels []Label
	payload       EventPayload
}

func NewBaseEvent(routesCapacity, labelsCapacity int) *BaseEvent {
	return &BaseEvent{
		routes:        make([]string, 0, routesCapacity),
		labels:        make([]Label, 0, labelsCapacity),
		produceLabels: make([]Label, 0, labelsCapacity),
		payload:       EventPayload{},
	}
}

func (b *BaseEvent) GetId() uint64 {
	return b.id
}

func (b *BaseEvent) SetId(id uint64) {
	b.id = id
}

func (b *BaseEvent) GetSourceId() uint64 {
	return b.sourceId
}

func (b *BaseEvent) SetSourceId(id uint64) {
	b.sourceId = id
}

func (b *BaseEvent) PushRoute(route string) {
	b.routes = append(b.routes, route)
}

func (b *BaseEvent) GetRoutes() []string {
	return b.routes
}

func (b *BaseEvent) SetRoutes(routes []string) {
	b.routes = routes
}

func (b *BaseEvent) CopyRoutes() []string {
	newRoutes := make([]string, len(b.routes))
	copy(newRoutes, b.routes)
	return newRoutes
}

func (b *BaseEvent) PopLabel() Label {
	label := b.labels[len(b.labels)-1]
	b.labels = b.labels[:len(b.labels)-1]
	return label
}

func (b *BaseEvent) PushLabel(label Label) {
	b.labels = append(b.labels, label)
}

func (b *BaseEvent) PushLabels(labels []Label) {
	for _, label := range labels {
		b.PushLabel(label)
	}
}

func (b *BaseEvent) SetLabels(labels []Label) {
	b.labels = labels
}

func (b *BaseEvent) GetLabels() []Label {
	return b.labels
}

func (b *BaseEvent) CopyLabels() []Label {
	newLabels := make([]Label, len(b.labels))
	copy(newLabels, b.labels)
	return newLabels
}

func (b *BaseEvent) GetLatestLabel() Label {
	if len(b.labels) == 0 {
		return "no labels exist"
	}
	return b.labels[len(b.labels)-1]
}

func (b *BaseEvent) PushProduceLabels() {
	for _, l := range b.produceLabels {
		b.PushLabel(l)
	}
	b.produceLabels = make([]Label, 0, len(b.produceLabels)) // FIXME
}

func (b *BaseEvent) GetProduceLabels() []Label {
	return b.produceLabels
}

func (b *BaseEvent) SetPayload(payload EventPayload) {
	b.payload = payload
}

func (b *BaseEvent) GetPayload() EventPayload {
	return b.payload
}

func (b *BaseEvent) Copy() Event {
	e := NewBaseEvent(len(b.routes), len(b.labels))
	newPayload := EventPayload{}
	for k, v := range b.payload {
		newPayload[k] = v // FIXME
	}
	e.payload = newPayload
	e.SetLabels(b.CopyLabels())
	e.SetRoutes(b.CopyRoutes())
	return e
}

type processors []Processor

func (ps processors) get(name string) (Processor, bool) {
	for _, processor := range ps {
		if processor.GetName() == name {
			return processor, true
		}
	}
	return nil, false
}

func (ps processors) getOrCreate(pType string) (Processor, error) {
	for _, processor := range ps {
		if processor.GetType() == pType {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("%s not found", pType)
}

func RegisterProcessorsFromFlow(dispatcher *Dispatcher, flow *Flow, factory ProcessorFactory) error {
	// FIXME return Job with processor struct and register outside

	var ps processors
	for _, pipeline := range flow.Pipelines {
		for i, job := range pipeline.Jobs {
			name := job.GetProcessorType() // FIXME check job processorName if exist
			var newProcessor Processor
			if p, ok := ps.get(name); ok {
				newProcessor = p
			} else {
				processor, err := factory.Create(job)
				if err != nil {
					return err
				}
				newProcessor = processor
			}

			var fullConsumeLabels []Label
			var fullProduceLabels []Label

			if len(job.Consume) == 0 {
				label := Label(job.ProcessorType)
				if job.Name != "" {
					label = Label(job.Name)
				}
				fullConsumeLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, label))
				fullConsumeLabels = append(fullConsumeLabels, fullConsumeLabel)
				if i == 0 {
					fullConsumeLabels = append(fullConsumeLabels, Label(pipeline.Name))
				}
			} else {
				for _, c := range job.Consume {
					if c == "init" { // FIXME
						fullConsumeLabels = append(fullConsumeLabels, c)
						continue
					}

					if _, ok := flow.getPipeline(string(c)); ok {
						fullConsumeLabels = append(fullConsumeLabels, c)
						continue
					}

					fullConsumeLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, c))
					fullConsumeLabels = append(fullConsumeLabels, fullConsumeLabel)
				}
			}

			if len(job.Produce) == 0 && i < (len(pipeline.Jobs)-1) {
				nextJob := pipeline.Jobs[i+1]
				label := Label(nextJob.ProcessorType)
				if nextJob.Name != "" {
					label = Label(nextJob.Name)
				}
				fullProduceLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, label))
				fullProduceLabels = append(fullProduceLabels, fullProduceLabel)
			} else {
				for _, p := range job.Produce {
					if _, ok := flow.getPipeline(string(p)); ok {
						fullProduceLabels = append(fullProduceLabels, p)
						continue
					}

					fullProduceLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, p))
					fullProduceLabels = append(fullProduceLabels, fullProduceLabel)
				}
			}

			dispatcher.RegisterTask(newProcessor.GetName(), fullConsumeLabels, fullProduceLabels, newProcessor) // FIXME name
		}
	}
	return nil
}
