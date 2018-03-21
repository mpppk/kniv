package twitter

import (
	"errors"
)

type Config struct {
	ScreenName        string
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	MaxTweetNum       int
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

	config.MaxTweetNum = 1000 // FIXME
	return config, nil
}
