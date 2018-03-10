package kniv

import "errors"

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
	urlEvent, ok := event.(*URLEvent)
	if !ok {
		return []Event{}, errors.New("invalid dispatched event found in ImageDownloadProcessor") // FIXME
	}
	_, err := Download(urlEvent.Url, urlEvent.Group)
	return []Event{}, err // FIXME
}
