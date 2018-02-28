package kniv

type Crawler interface {
	SetResourceChannel(chan string)
	SendResourceUrlsToChannel()
}

var Crawlers []Crawler

func RegisterCrawler(crawler Crawler) {
	Crawlers = append(Crawlers, crawler)
}
