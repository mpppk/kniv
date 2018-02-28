package kniv

type Crawler interface {
	SetResourceChannel(chan string)
	SendResourceUrlsToChannel()
}
