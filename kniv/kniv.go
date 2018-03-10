package kniv

import (
	"sync"
)

type ResourceType string
type Label string

type URLEvent struct {
	*BaseEvent
	Url   string
	Group string
}

func NewURLEvent(url, group string, routesCapacity, labelsCapacity int) *URLEvent {
	return &URLEvent{
		BaseEvent: NewBaseEvent(routesCapacity, labelsCapacity),
		Url:       url,
		Group:     group,
	}
}

type Event interface {
	PushRoute(route string)
	GetRoutes() []string
	PopLabel() Label
	PushLabel(label Label)
	GetLatestLabel() Label
	GetLabels() []Label
}

type BaseEvent struct {
	routes []string
	labels []Label
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
