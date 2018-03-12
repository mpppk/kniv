package twitter

import (
	"errors"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/mpppk/kniv/kniv"
	"log"
	"net/url"
	"sync"
	"time"
)

type Config struct {
	ScreenName        string
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	SinceDate         time.Time
}

type Crawler struct {
	client          *anaconda.TwitterApi
	config          *Config
	resourceChannel chan kniv.Event
	rootDownloadDir string
}

func CreateClient(config *Config) *anaconda.TwitterApi {
	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)

	api := anaconda.NewTwitterApi(config.AccessToken, config.AccessTokenSecret)
	api.SetLogger(anaconda.BasicLogger) // logger を設定
	return api
}

func NewCrawler(config *Config) kniv.Crawler {
	client := CreateClient(config)

	return &Crawler{
		client: client,
		config: config,
	}
}

func (c *Crawler) SetResourceChannel(q chan kniv.Event) {
	c.resourceChannel = q
}

func (c *Crawler) SetRootDownloadDir(dir string) {
	c.rootDownloadDir = dir
}

func (c *Crawler) StartResourceSending(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("start resource sending in twitter")
	tweets, err := c.Fetch(0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("tweets num")
	fmt.Println(len(tweets))

	for _, tweet := range tweets {
		for _, media := range tweet.Entities.Media {
			fmt.Println(media.Media_url)
		}
	}
}

func (c *Crawler) Fetch(offset, limit int) ([]anaconda.Tweet, error) {
	values := url.Values{
		"screen_name":     []string{c.config.ScreenName},
		"count":           []string{"200"},
		"exclude_replies": []string{"true"},
		"trim_user":       []string{"true"},
		"include_rts":     []string{"false"},
	}

	return c.client.GetUserTimeline(values)
}

func toConfig(configMap map[string]interface{}) (*Config, error) {
	config := &Config{}

	if consumerKey, ok := configMap["consumer_key"].(string); ok {
		config.ConsumerKey = consumerKey
	} else {
		return nil, errors.New("consumer_key not found in setting file")
	}

	if consumerSecret, ok := configMap["consumer_secret"].(string); ok {
		config.ConsumerSecret = consumerSecret
	} else {
		return nil, errors.New("consumer_secret not found in setting file")
	}

	if accessToken, ok := configMap["access_token"].(string); ok {
		config.AccessToken = accessToken
	} else {
		return nil, errors.New("access_token not found in setting file")
	}

	if accessTokenSecret, ok := configMap["access_token_secret"].(string); ok {
		config.AccessTokenSecret = accessTokenSecret
	} else {
		return nil, errors.New("access_token_secret not found in setting file")
	}
	return config, nil
}

type CrawlerFactory struct{}

func (c *CrawlerFactory) Create(crawlersSetting map[string]interface{}) (kniv.Crawler, error) {

	setting, ok := crawlersSetting["twitter"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid setting in tumblr key")
	}

	config, err := toConfig(setting)
	return NewCrawler(config), err
}
