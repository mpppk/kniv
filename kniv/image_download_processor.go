package kniv

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

	_, err = Download(eventUrl, group)
	return []Event{}, err // FIXME
}
