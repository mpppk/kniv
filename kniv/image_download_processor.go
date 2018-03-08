package kniv

type ImageDownloadProcessor struct {
	*BaseProcessor
	rootDestination string
}

func NewImageDownloadProcessor(queueSize int, rootDestination string) *ImageDownloadProcessor {
	return &ImageDownloadProcessor{
		BaseProcessor: &BaseProcessor{
			Name:    "image download processor",
			inChan:  make(chan Resource, queueSize),
			Process: DownloadFromResource,
		},
		rootDestination: rootDestination,
	}
}

func DownloadFromResource(resource Resource) ([]Resource, error) {
	_, err := Download(resource.Url, resource.DstPath)
	return []Resource{{ResourceType: resource.NextResourceType}}, err
}
