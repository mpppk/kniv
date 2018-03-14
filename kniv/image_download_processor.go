package kniv

import (
	"path"
)

type DownloadResultEvent struct {
	*BaseEvent
	Success bool
}

type ImageDownloadProcessor struct {
	*BaseProcessor
	rootDestination string
}

func NewImageDownloadProcessor(queueSize int, rootDestination string) *ImageDownloadProcessor {
	return &ImageDownloadProcessor{
		BaseProcessor: &BaseProcessor{
			Name:    "downloader",
			inChan:  make(chan Event, queueSize),
			Process: DownloadFromResource,
		},
		rootDestination: rootDestination,
	}
}

func DownloadFromResource(event Event) ([]Event, error) {
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

	downloadPath := path.Join(group, user)
	downloaded, err := Download(eventUrl, downloadPath)
	event.GetPayload()["downloaded"] = downloaded
	event.GetPayload()["download_path"] = downloadPath // FIXME use root dir and user
	return []Event{}, err
}
