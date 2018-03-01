package kniv

type Resource struct {
	ResourceType string
	Url          string
	DstPath      string
}

type Crawler interface {
	SetResourceChannel(chan Resource)
	SendResourceUrlsToChannel()
}

var Crawlers []Crawler

func RegisterCrawler(crawler Crawler) {
	Crawlers = append(Crawlers, crawler)
}
