package downloader

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mpppk/kniv/etc"
)

func FetchURL(wg *sync.WaitGroup, q chan string, dstDir string, sleepMilliSec time.Duration) {
	defer wg.Done()
	queueSize := 0
	for {
		fileUrl, ok := <-q // closeされると ok が false になる
		if !ok {
			fmt.Println("url fetching is terminated")
			return
		}

		if len(q) != queueSize {
			queueSize = len(q)
			log.Printf("current URL queue size: %d\n", queueSize)
		}

		_, err := img.Download(fileUrl, dstDir)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(sleepMilliSec * time.Millisecond)
	}
}
