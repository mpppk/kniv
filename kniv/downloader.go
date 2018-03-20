package kniv

import (
	"path"
)

type Downloader struct {
	*BaseProcessor
	rootDestination string
}

func NewDownloader(queueSize int, rootDestination string) *Downloader {
	downloader := &Downloader{
		BaseProcessor: &BaseProcessor{
			Name:   "downloader",
			inChan: make(chan Event, queueSize),
		},
		rootDestination: rootDestination,
	}
	downloader.BaseProcessor.Process = downloader.DownloadFromResource
	return downloader
}

func (p *Downloader) DownloadFromResource(event Event) ([]Event, error) {
	eventUrl, err := event.GetPayload().GetString("url")
	if err != nil {
		return []Event{}, err // FIXME
	}

	group, err := event.GetPayload().GetString("group")
	if err != nil {
		return []Event{}, err // FIXME
	}

	user, err := event.GetPayload().GetString("user")
	if err != nil {
		return []Event{}, err // FIXME
	}

	downloadPath := path.Join(p.rootDestination, group, user)
	downloaded, err := Download(eventUrl, downloadPath)
	event.GetPayload()["downloaded"] = downloaded
	event.GetPayload()["download_path"] = downloadPath
	return []Event{event}, err
}
