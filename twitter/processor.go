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

func (c *Processor) Fetch(maxId int64, count int) ([]anaconda.Tweet, error) {
	values := url.Values{
		"screen_name":     []string{c.config.ScreenName},
		"count":           []string{fmt.Sprint(count)},
		"exclude_replies": []string{"true"},
		"trim_user":       []string{"true"},
		"include_rts":     []string{"false"},
	}

	if maxId > 0 {
		values["max_id"] = []string{fmt.Sprint(maxId - 1)} // minus 1 because max_id includes specified id tweet
	}

	return c.client.GetUserTimeline(values)
}

func (c *Processor) Process(event kniv.Event) ([]kniv.Event, error) {
	payload := event.GetPayload()
	sinceIdPayload, ok := payload["since_id"]
	if !ok {
		sinceIdPayload = nil
	}
	var sinceId int64
	switch v := sinceIdPayload.(type) {
	case int:
		sinceId = int64(v)
	case int64:
		sinceId = v
	case float64:
		sinceId = int64(v)
	case nil:
		sinceId = -1
	}

	countPayload, ok := payload["count"]
	if !ok {
		log.Fatal("count key not found in payload")
	}
	var count int
	switch v := countPayload.(type) {
	case int:
		count = v
	case float64:
		count = int(v)
	}

	tweetNum := 0
	for {
		if tweetNum >= c.config.MaxTweetNum {
			return []kniv.Event{}, nil
		}

		tweets, err := c.Fetch(sinceId, count)
		if err != nil {
			return nil, err
		}

		var nextSinceId int64
		nextSinceId = -1
		var events []kniv.Event
		for _, tweet := range tweets {
			for _, media := range tweet.Entities.Media {
				r := kniv.NewBaseEvent(10, 10) // FIXME
				r.GetPayload()["url"] = media.Media_url
				r.GetPayload()["group"] = path.Join("twitter", c.config.ScreenName) // FIXME
				r.GetPayload()["count"] = count
				r.GetPayload()["user"] = c.config.ScreenName
				events = append(events, r)
			}
			if tweet.Id < nextSinceId || nextSinceId < 0 {
				nextSinceId = tweet.Id
			}
			tweetNum++
		}

		if len(events) > 0 {
			for _, e := range events {
				e.GetPayload()["since_id"] = nextSinceId
			}

			return events, nil
		}
	}
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

func CreateClient(config *Config) *anaconda.TwitterApi {
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)

	api := anaconda.NewTwitterApi(config.AccessToken, config.AccessTokenSecret)
	api.SetLogger(anaconda.BasicLogger) // logger を設定
	return api
}
