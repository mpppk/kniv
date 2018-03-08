package twitter

import (
	"fmt"
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

func (c *Processor) Process(resource kniv.Resource) ([]kniv.Resource, error) {
	tweets, err := c.Fetch(0, 10)
	if err != nil {
		return nil, err
	}

	var resources []kniv.Resource
	for _, tweet := range tweets {
		for _, media := range tweet.Entities.Media {
			r := kniv.Resource{
				ResourceType:     "twitter.image",
				NextResourceType: "end", // TODO: Add NextResourceType by dispatcher
				Url:              media.Media_url,
				DstPath:          path.Join("twitter", c.config.ScreenName),
			}
			resources = append(resources, r)
			fmt.Println(media.Media_url)
		}
	}
	return resources, nil
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
	processor.BaseProcessor.Process = processor.Process
	return processor
}