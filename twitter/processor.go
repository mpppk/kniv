package twitter

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/mpppk/kniv/kniv"
	"net/url"
	"path"
)

type Processor struct {
	*kniv.BaseProcessor
	client *anaconda.TwitterApi
	config *Config
}

func (c *Processor) Fetch(offset, limit int) ([]anaconda.Tweet, error) {
	values := url.Values{
		"screen_name":     []string{c.config.ScreenName},
		"count":           []string{"200"},
		"exclude_replies": []string{"true"},
		"trim_user":       []string{"true"},
		"include_rts":     []string{"false"},
	}

	return c.client.GetUserTimeline(values)
}

func (c *Processor) Process(event kniv.Event) ([]kniv.Event, error) {
	offset := 0 // FIXME
	limit := 10 // FIXME
	tweets, err := c.Fetch(offset, limit)
	if err != nil {
		return nil, err
	}

	var events []kniv.Event
	for _, tweet := range tweets {
		for _, media := range tweet.Entities.Media {
			r := kniv.NewBaseEvent(10, 10)
			r.GetPayload()["url"] = media.Media_url
			r.GetPayload()["group"] = path.Join("twitter", c.config.ScreenName) // FIXME
			r.GetPayload()["offset"] = offset
			r.GetPayload()["limit"] = limit
			r.GetPayload()["user"] = c.config.ScreenName
			events = append(events, r)
		}
	}
	return events, nil
}

func NewProcessorFromConfigMap(queueSize int, configMap map[string]interface{}) (Processor, error) {
	config, err := toConfig(configMap)
	return NewProcessor(queueSize, config), err
}

func NewProcessor(queueSize int, config *Config) Processor {
	processor := Processor{
		BaseProcessor: kniv.NewBaseProcessor(queueSize),
		client:        CreateClient(config),
		config:        config,
	}
	processor.BaseProcessor.Name = "twitter"
	processor.BaseProcessor.Process = processor.Process
	return processor
}
