package downloader

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mpppk/kniv/etc"
	"github.com/mpppk/kniv/kniv"
)

type Downloader struct {
	Channel       chan string
	wg            sync.WaitGroup
	sleepMilliSec time.Duration
}

func New(queueSize int, sleepMilliSec time.Duration) *Downloader {
	return &Downloader{
		Channel:       make(chan string, queueSize),
		sleepMilliSec: sleepMilliSec,
	}
}

func (d *Downloader) FetchURL(dstDir string) {
	defer d.wg.Done()
	queueSize := 0
	for {
		fileUrl, ok := <-d.Channel // closeされると ok が false になる
		if !ok {
			fmt.Println("url fetching is terminated")
			return
		}

		if len(d.Channel) != queueSize {
			queueSize = len(d.Channel)
			log.Printf("current URL queue size: %d\n", queueSize)
		}

		_, err := img.Download(fileUrl, dstDir)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(d.sleepMilliSec * time.Millisecond)
	}
}

func (d *Downloader) RegisterCrawler(crawler kniv.Crawler, dstDir string) {
	//	TODO: crawler.GetDownloadDestinationsを実装(downloader側で登録するのに使う)

	d.wg.Add(1)
	crawler.SetResourceChannel(d.Channel)
	go d.FetchURL(dstDir)
}

func (d *Downloader) SetDownloadDestination(crawler kniv.Crawler, dstDir string) {

}
