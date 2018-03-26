package kniv

import (
	"github.com/mitchellh/mapstructure"
	"path"
)

type Downloader struct {
	*BaseProcessor
	rootDestination string
}

type DownloaderArgs struct {
	BaseArgs
	RootDestination string
	// FIXME interval, etc...
}

const downloaderType = "downloader"

func NewDownloader(args *DownloaderArgs) *Downloader {
	downloader := &Downloader{
		BaseProcessor: &BaseProcessor{
			Name:   downloaderType,
			inChan: make(chan Event, args.QueueSize),
		},
		rootDestination: args.RootDestination,
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

type DownloaderGenerator struct{}

func (g *DownloaderGenerator) Generate(intfArgs interface{}) (Processor, error) {
	var args DownloaderArgs
	err := mapstructure.Decode(intfArgs, &args)
	if err != nil {
		return nil, err
	}
	return NewDownloader(&args), nil
}

func (g *DownloaderGenerator) GetType() string {
	return downloaderType
}
