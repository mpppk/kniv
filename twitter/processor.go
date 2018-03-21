package twitter

import (
	"net/url"
	"path"

	"log"

	"fmt"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mpppk/kniv/kniv"
)

type Processor struct {
	*kniv.BaseProcessor
	client *anaconda.TwitterApi
	config *Config
}

func (c *Processor) Fetch(offset, limit int) ([]anaconda.Tweet, error) {
	// FIXME use offset
	values := url.Values{
		"screen_name":     []string{c.config.ScreenName},
		"count":           []string{fmt.Sprint(limit)},
		"exclude_replies": []string{"true"},
		"trim_user":       []string{"true"},
		"include_rts":     []string{"false"},
	}

	return c.client.GetUserTimeline(values)
}

func (c *Processor) Process(event kniv.Event) ([]kniv.Event, error) {
	payload := event.GetPayload()
	offsetPayload, ok := payload["offset"]
	if !ok {
		log.Fatal("offset key not found in payload")
	}
	var offset int
	switch v := offsetPayload.(type) {
	case int:
		offset = v
	case float64:
		offset = int(v)
	}

	limitPayload, ok := payload["limit"]
	if !ok {
		log.Fatal("limit key not found in payload")
	}
	var limit int
	switch v := limitPayload.(type) {
	case int:
		limit = v
	case float64:
		limit = int(v)
	}

	if limit > c.config.MaxOffset {
		return []kniv.Event{}, nil
	}

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
