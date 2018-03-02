package downloader

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mpppk/kniv/kniv"
	"path"
)

type Downloader struct {
	Channel         chan kniv.Resource
	wg              sync.WaitGroup
	sleepMilliSec   time.Duration
	rootDestination string
}

func New(queueSize int, sleepMilliSec time.Duration, rootDestination string) *Downloader {
	return &Downloader{
		Channel:         make(chan kniv.Resource, queueSize),
		sleepMilliSec:   sleepMilliSec,
		rootDestination: rootDestination,
	}
}

func (d *Downloader) Start() {
	defer d.wg.Done()
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

		_, err := Download(resource.Url, path.Join(d.rootDestination, resource.DstPath))
		if err != nil {
			log.Println(err)
		}
		time.Sleep(d.sleepMilliSec * time.Millisecond)
	}
}

func (d *Downloader) RegisterCrawler(crawler kniv.Crawler) {
	d.wg.Add(1)
	crawler.SetResourceChannel(d.Channel)
	go d.Start()
}

func (d *Downloader) SetDownloadDestination(crawler kniv.CrawlerFactory, dstDir string) {

}
