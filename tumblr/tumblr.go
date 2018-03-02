package tumblr

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"

	"os"

	"errors"
	"github.com/MariaTerzieva/gotumblr"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/mpppk/kniv/downloader"
	"github.com/mpppk/kniv/kniv"
	"github.com/skratchdot/open-golang/open"
	"time"
)

type VideoPost struct {
	gotumblr.BasePost
	VideoUrl string `json:"video_url"`
}

func filterExistFileUrls(fileUrls []string, dir string) (filteredFileUrls []string, err error) {
	for _, fileUrl := range fileUrls {
		exist, err := isExistFileUrl(fileUrl, dir)
		if err != nil {
			return nil, err
		}

		if !exist {
			filteredFileUrls = append(filteredFileUrls, fileUrl)
		}
	}
	return filteredFileUrls, err
}

func isExistFileUrl(fileUrl string, dir string) (bool, error) {
	fileName, err := downloader.GetFileNameFromUrl(fileUrl)
	if err != nil {
		return false, err
	}

	if !downloader.IsExist(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return false, err
		}
	}

	if downloader.IsExist(path.Join(dir, fileName)) {
		return true, nil
	}
	return false, nil
}

func convertJsonToVideoPosts(jsonPosts []json.RawMessage) []VideoPost {
	var videoPosts []VideoPost
	//var videoPost gotumblr.VideoPost
	var videoPost VideoPost
	for _, post := range jsonPosts {
		//fmt.Println(fmt.Sprintf("%s", post))
		json.Unmarshal(post, &videoPost)
		if videoPost.PostType != "video" {
			continue
		}
		videoPosts = append(videoPosts, videoPost)
	}
	return videoPosts
}

func convertJsonToPhotoPosts(jsonPosts []json.RawMessage) []gotumblr.PhotoPost {
	var photoPosts []gotumblr.PhotoPost
	var photoPost gotumblr.PhotoPost
	for _, post := range jsonPosts {
		json.Unmarshal(post, &photoPost)
		if photoPost.PostType != "photo" {
			continue
		}
		photoPosts = append(photoPosts, photoPost)
	}
	return photoPosts
}

func getVideoUrlsFromAPIResponse(videoPosts []VideoPost) []string {
	var videoUrls []string
	for _, post := range videoPosts {
		if post.PostType != "video" {
			continue
		}
		videoUrls = append(videoUrls, post.VideoUrl)
	}
	return videoUrls
}

func getImageUrlsFromAPIResponse(photoPosts []gotumblr.PhotoPost) []string {
	var photoUrls []string
	for _, post := range photoPosts {
		if post.PostType != "photo" {
			continue
		}

		for _, photo := range post.Photos {
			maxSizeUrl := getMaxSizeUrl(photo)
			photoUrls = append(photoUrls, maxSizeUrl)
		}
	}
	return photoUrls
}

func getMaxSizeUrl(photo gotumblr.PhotoObject) string {
	maxSize := photo.Alt_sizes[0]
	for _, size := range photo.Alt_sizes {
		if maxSize.Height < size.Height {
			maxSize = size
		}
	}
	return maxSize.Url
}

func authorize() {
	oauthClient := &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  "xzORqsOREcMl19OIQjbgl3pBzfqlYUrqU4LzwZLkCEkqt2baSE",
			Secret: "8xOEM1eThFDtkDyluDK5wKZK9LBn3Cm8l5wzuR0dZTdXRNaFWm",
		},
		TemporaryCredentialRequestURI: "http://www.tumblr.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "http://www.tumblr.com/oauth/authorize",
		TokenRequestURI:               "http://www.tumblr.com/oauth/access_token",
	}

	scope := url.Values{"scope": {"read_public,write_public,read_private,write_private"}}

	tempCredentials, err := oauthClient.RequestTemporaryCredentials(nil, "", scope)
	if err != nil {
		log.Fatal("RequestTemporaryCredentials:", err)
	}

	u := oauthClient.AuthorizationURL(tempCredentials, nil)
	fmt.Printf("1. Go to %s\n2. Authorize the application\n3. Enter verification code:\n", u)
	open.Run(u)

	var code string
	fmt.Scanln(&code)

	fmt.Println("InputCode: ", code)

	tokenCard, _, err := oauthClient.RequestToken(nil, tempCredentials, code)
	if err != nil {
		log.Fatal("RequestToken:", err)
	}

	fmt.Println("Token: ", tokenCard.Token)
	fmt.Println("Secret: ", tokenCard.Secret)
}

type CrawlerFactory struct{}

func (c *CrawlerFactory) Create(crawlersSetting map[string]interface{}) (kniv.Crawler, error) {

	optMap, ok := crawlersSetting["tumblr"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid setting in tumblr key")
	}

	opt, err := toOpt(optMap)
	return NewCrawler(opt), err
}

func toOpt(optMap map[string]interface{}) (*Opt, error) {
	// TODO: TEST
	opt := &Opt{
		Offset:                   0,
		MaxBlogNum:               1000,
		PostNumPerBlog:           20,
		APIIntervalMilliSec:      time.Duration(10000),
		DownloadIntervalMilliSec: time.Duration(10000),
		DstDirMap: map[string]string{
			"photo": "tumblr/photo",
			"video": "tumblr/video",
		},
	}

	if consumerKey, ok := optMap["consumer_key"].(string); ok {
		opt.ConsumerKey = consumerKey
	} else {
		return nil, errors.New("consumer_key not found in setting file")
	}

	if consumerSecret, ok := optMap["consumer_secret"].(string); ok {
		opt.ConsumerSecret = consumerSecret
	} else {
		return nil, errors.New("consumer_secret not found in setting file")
	}

	if oauthToken, ok := optMap["oauth_token"].(string); ok {
		opt.OauthToken = oauthToken
	} else {
		return nil, errors.New("oauth_token not found in setting file")
	}

	if oauthSecret, ok := optMap["oauth_secret"].(string); ok {
		opt.OauthSecret = oauthSecret
	} else {
		return nil, errors.New("oauth_secret not found in setting file")
	}

	if offset, ok := optMap["offset"].(int); ok {
		opt.Offset = offset
	}

	if maxBlogNum, ok := optMap["max_blog_num"].(int); ok {
		opt.MaxBlogNum = maxBlogNum
	}

	if postNumPerBlog, ok := optMap["post_num_per_blog"].(int); ok {
		opt.PostNumPerBlog = postNumPerBlog
	}

	if apiInterval, ok := optMap["api_interval_millisec"].(int); ok {
		opt.APIIntervalMilliSec = time.Duration(apiInterval) * time.Millisecond
	}

	if downloadInterval, ok := optMap["download_interval_millisec"].(int); ok {
		opt.DownloadIntervalMilliSec = time.Duration(downloadInterval) * time.Millisecond
	}
	return opt, nil
}

func init() {
	kniv.RegisterCrawlerFactory(&CrawlerFactory{})
}
