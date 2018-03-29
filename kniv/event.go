package kniv

import (
	"errors"
)

type Label string

type Labels []Label

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

func (ls *Labels) Pop() Label {
	label := (*ls)[len(*ls)-1]
	*ls = (*ls)[:len(*ls)-1]
	return label
}

func (ls *Labels) Push(label Label) {
	*ls = append(*ls, label)
}

func (ls *Labels) PushAll(labels *Labels) {
	for _, label := range *labels {
		ls.Push(label)
	}
}

func (ls *Labels) Copy() *Labels {
	newLabels := make(Labels, len(*ls))
	copy(newLabels, *ls)
	return &newLabels
}

func (ls *Labels) GetLatest() Label {
	if len(*ls) == 0 {
		return "no labels exist"
	}
	return (*ls)[len(*ls)-1]
}

type Event interface {
	GetId() uint64
	SetId(uint64)
	GetSourceId() uint64
	SetSourceId(uint64)
	PushRoute(route string)
	//GetRoutes() []string
	SetRoutes([]string)
	CopyRoutes() []string
	//PopLabel() Label
	//PushLabel(label Label)
	//PushLabels(labels Labels)
	//SetLabels(labels Labels)
	//CopyLabels() Labels
	//GetProduceLabels() Labels
	//PushProduceLabels()
	//GetLatestLabel() Label
	GetLabels() *Labels
	SetLabels(labels *Labels)
	SetPayload(payload EventPayload)
	GetPayload() EventPayload
	Copy() Event
}

type BaseEvent struct {
	id       uint64
	sourceId uint64
	routes   []string
	labels   *Labels
	//produceLabels *Labels
	payload EventPayload
}

func NewBaseEvent(labelsCapacity, routesCapacity int) *BaseEvent {
	labels := make(Labels, 0, labelsCapacity)
	return &BaseEvent{
		routes: make([]string, 0, routesCapacity),
		//labels:        make(Label, 0, labelsCapacity),
		labels: &labels,
		//produceLabels: make([]Label, 0, labelsCapacity),
		payload: EventPayload{},
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

func (b *BaseEvent) GetLabels() *Labels {
	return b.labels
}

func (b *BaseEvent) SetLabels(labels *Labels) {
	b.labels = labels
}

//func (b *BaseEvent) PushProduceLabels() {
//	for _, l := range b.produceLabels {
//		b.GetLabels().Push(l)
//	}
//	b.produceLabels = make([]Label, 0, len(b.produceLabels)) // FIXME
//}
//
//func (b *BaseEvent) GetProduceLabels() []Label {
//	return b.produceLabels
//}

func (b *BaseEvent) SetPayload(payload EventPayload) {
	b.payload = payload
}

func (b *BaseEvent) GetPayload() EventPayload {
	return b.payload
}

func (b *BaseEvent) Copy() Event {
	e := NewBaseEvent(len(b.routes), len(*b.labels))
	newPayload := EventPayload{}
	for k, v := range b.payload {
		newPayload[k] = v // FIXME
	}
	e.payload = newPayload
	e.SetLabels(b.GetLabels().Copy())
	e.SetRoutes(b.CopyRoutes())
	return e
}
