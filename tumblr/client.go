package tumblr

import (
	"fmt"
	"time"

	"log"

	"github.com/MariaTerzieva/gotumblr"
)

type Opt struct {
	ConsumerKey              string
	ConsumerSecret           string
	OauthToken               string
	OauthSecret              string
	Offset                   int
	MaxBlogNum               int
	PostNumPerBlog           int
	APIIntervalMilliSec      time.Duration
	DownloadIntervalMilliSec time.Duration
	DstDirMap                map[string]string
	DownloadQueueSize        int
}

type Client struct {
	*gotumblr.TumblrRestClient
	opt           *Opt
	photoFetchNum int
	videoFetchNum int
	photoDstDir   string
	videoDstDir   string
	URLChannel    chan string
}

func NewClient(opt *Opt) *Client {
	rawClient := gotumblr.NewTumblrRestClient(
		opt.ConsumerKey,
		opt.ConsumerSecret,
		opt.OauthToken,
		opt.OauthSecret,
		"callback_url",
		"http://api.tumblr.com",
	)

	photoDstDir := "photos"
	if v, ok := opt.DstDirMap["photo"]; ok {
		photoDstDir = v
	}
	videoDstDir := "videos"
	if v, ok := opt.DstDirMap["video"]; ok {
		videoDstDir = v
	}

	return &Client{
		TumblrRestClient: rawClient,
		opt:              opt,
		photoFetchNum:    opt.Offset,
		videoFetchNum:    opt.Offset,
		photoDstDir:      photoDstDir,
		videoDstDir:      videoDstDir,
		URLChannel:       make(chan string, opt.DownloadQueueSize),
	}
}

func (c *Client) GetPhotoUrls(blogName string, apiOffset int) []string {
	apiOpt := map[string]string{"offset": fmt.Sprint(apiOffset)}
	photoRes := c.Posts(blogName, "photo", apiOpt)
	return getImageUrlsFromAPIResponse(convertJsonToPhotoPosts(photoRes.Posts))
}

func (c *Client) GetVideoUrls(blogName string, apiOffset int) []string {
	apiOpt := map[string]string{"offset": fmt.Sprint(apiOffset)}
	videoRes := c.Posts(blogName, "video", apiOpt)
	return getVideoUrlsFromAPIResponse(convertJsonToVideoPosts(videoRes.Posts))
}

func (c *Client) GetBlogNames(max int) []string {
	var blogNames []string
	offset := 0
	for offset <= max {
		blogs := c.Following(map[string]string{"offset": fmt.Sprint(offset)}).Blogs

		if len(blogs) == 0 {
			fmt.Println("blog num zero")
			break
		}
		for _, blog := range blogs {
			blogNames = append(blogNames, blog.Name)
		}
		offset += 20
	}
	return blogNames
}

func (c *Client) sendVideoURLsToChannel(blogName string) {
	c.sendFileURLsToChannel(c.GetVideoUrls, blogName, c.videoDstDir)
}

func (c *Client) sendPhotoURLsToChannel(blogName string) {
	c.sendFileURLsToChannel(c.GetPhotoUrls, blogName, c.photoDstDir)
}

func (c *Client) sendFileURLsToChannel(getFileUrls func(string, int) []string, blogName, dstDir string) {
	fetchNum := c.opt.Offset
	for fetchNum <= c.opt.PostNumPerBlog+c.opt.Offset {
		fileUrls := getFileUrls(blogName, fetchNum)

		log.Printf("%d URLs are found on %s %d-%d / %d",
			len(fileUrls), blogName, fetchNum, fetchNum+20, c.opt.PostNumPerBlog+c.opt.Offset)

		fileUrls, err := filterExistFileUrls(fileUrls, dstDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, fileUrl := range fileUrls {
			c.URLChannel <- fileUrl
		}

		time.Sleep(c.opt.APIIntervalMilliSec * time.Millisecond)

		if len(fileUrls) == 0 {
			return
		}
		fetchNum += 20
	}
}
