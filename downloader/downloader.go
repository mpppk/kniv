package downloader

import (
	"fmt"
	"log"
	"time"

	"github.com/mpppk/kniv/kniv"
	"sync"
)

type Downloader struct {
	Channel         chan kniv.Resource
	sleepMilliSec   time.Duration
	rootDestination string
	crawlers        []kniv.Crawler
	wg              *sync.WaitGroup
}

func New(queueSize int, sleepMilliSec time.Duration) *Downloader {
	return &Downloader{
		Channel:       make(chan kniv.Resource, queueSize),
		sleepMilliSec: sleepMilliSec,
		wg:            &sync.WaitGroup{},
	}
}

func (d *Downloader) WatchResource() {
	queueSize := 0
	for {
		resource, ok := <-d.Channel // closeされると ok が false になる
		if !ok {
			fmt.Println("url fetching is terminated")
			return
		}

		if len(d.Channel) != queueSize {
			queueSize = len(d.Channel)
			log.Printf("current URL queue size: %d\n", queueSize)
		}

		_, err := Download(resource.Url, resource.DstPath)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(d.sleepMilliSec * time.Millisecond)
	}
}

func (d *Downloader) RegisterCrawler(crawler kniv.Crawler) {
	crawler.SetResourceChannel(d.Channel)
	d.crawlers = append(d.crawlers, crawler)
}

func (d *Downloader) StartCrawl() {
	go d.WatchResource()
	for _, crawler := range d.crawlers {
		d.wg.Add(1)
		go crawler.StartResourceSending(d.wg)
	}
	d.wg.Wait()
	close(d.Channel)
}
