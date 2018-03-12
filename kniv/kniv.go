package kniv

import (
	"errors"
	"sync"
)

type ResourceType string
type Label string

type URLEvent struct {
	*BaseEvent
}

func NewURLEvent(url, group string, routesCapacity, labelsCapacity int) *URLEvent {
	payload := map[string]interface{}{
		"url":   url,
		"group": group,
	}
	urlEvent := &URLEvent{
		BaseEvent: NewBaseEvent(routesCapacity, labelsCapacity),
	}
	urlEvent.SetPayload(payload)
	return urlEvent
}

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

type Event interface {
	PushRoute(route string)
	GetRoutes() []string
	PopLabel() Label
	PushLabel(label Label)
	GetLatestLabel() Label
	GetLabels() []Label
	SetPayload(payload EventPayload)
	GetPayload() EventPayload
}

type BaseEvent struct {
	routes  []string
	labels  []Label
	payload EventPayload
}

func NewBaseEvent(routesCapacity, labelsCapacity int) *BaseEvent {
	return &BaseEvent{
		routes: make([]string, 0, routesCapacity),
		labels: make([]Label, 0, labelsCapacity),
	}
}

func (b *BaseEvent) PushRoute(route string) {
	b.routes = append(b.routes, route)
}

func (b *BaseEvent) GetRoutes() []string {
	return b.routes
}

func (b *BaseEvent) PopLabel() Label {
	label := b.labels[0]
	b.labels = b.labels[1:]
	return label
}

func (b *BaseEvent) PushLabel(label Label) {
	b.labels = append(b.labels, label)
}

func (b *BaseEvent) GetLabels() []Label {
	return b.labels
}

func (b *BaseEvent) GetLatestLabel() Label {
	if len(b.labels) == 0 {
		return "no labels exist"
	}
	return b.labels[len(b.labels)-1]
}

func (b *BaseEvent) SetPayload(payload EventPayload) {
	b.payload = payload
}

func (b *BaseEvent) GetPayload() EventPayload {
	return b.payload
}

type Crawler interface {
	SetResourceChannel(chan URLEvent)
	SetRootDownloadDir(dir string)
	StartResourceSending(wg *sync.WaitGroup)
}

type CrawlerFactory interface {
	Create(crawlersSetting map[string]interface{}) (Crawler, error)
}

var CrawlerFactories []CrawlerFactory

func RegisterCrawlerFactory(crawlerGenerator CrawlerFactory) {
	CrawlerFactories = append(CrawlerFactories, crawlerGenerator)
}
