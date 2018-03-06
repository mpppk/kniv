package kniv

import "sync"

type ResourceType string

type Resource struct {
	ResourceType     ResourceType
	NextResourceType ResourceType
	Url              string
	DstPath          string
}

type Crawler interface {
	SetResourceChannel(chan Resource)
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
