package tumblr

import (
	"fmt"
	"time"

	"log"

	"github.com/MariaTerzieva/gotumblr"
	"github.com/mpppk/kniv/kniv"
	"path"
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

type Crawler struct {
	*gotumblr.TumblrRestClient
	opt             *Opt
	photoFetchNum   int
	videoFetchNum   int
	photoDstDir     string
	videoDstDir     string
	resourceChannel chan kniv.Resource
}

func NewCrawler(opt *Opt) kniv.Crawler {
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

	return &Crawler{
		TumblrRestClient: rawClient,
		opt:              opt,
		photoFetchNum:    opt.Offset,
		videoFetchNum:    opt.Offset,
		photoDstDir:      photoDstDir,
		videoDstDir:      videoDstDir,
	}
}

func (c *Crawler) SetResourceChannel(q chan kniv.Resource) {
	c.resourceChannel = q
}

func (c *Crawler) SendResourceUrlsToChannel() {
	blogNames := c.getBlogNames(c.opt.MaxBlogNum)
	for i, blogName := range blogNames {
		fmt.Printf("---- fetch from %s %d/%d----\n", blogName, i, len(blogNames))
		c.sendPhotoURLsToChannel(blogName)
		c.sendVideoURLsToChannel(blogName)
	}
	close(c.resourceChannel)
}

func (c *Crawler) getPhotoUrls(blogName string, apiOffset int) []string {
	apiOpt := map[string]string{"offset": fmt.Sprint(apiOffset)}
	photoRes := c.Posts(blogName, "photo", apiOpt)
	return getImageUrlsFromAPIResponse(convertJsonToPhotoPosts(photoRes.Posts))
}

func (c *Crawler) getVideoUrls(blogName string, apiOffset int) []string {
	apiOpt := map[string]string{"offset": fmt.Sprint(apiOffset)}
	videoRes := c.Posts(blogName, "video", apiOpt)
	return getVideoUrlsFromAPIResponse(convertJsonToVideoPosts(videoRes.Posts))
}

func (c *Crawler) getBlogNames(max int) []string {
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

func (c *Crawler) sendVideoURLsToChannel(blogName string) {
	c.sendFileURLsToChannel(c.getVideoUrls, blogName, c.videoDstDir)
}

func (c *Crawler) sendPhotoURLsToChannel(blogName string) {
	c.sendFileURLsToChannel(c.getPhotoUrls, blogName, c.photoDstDir)
}

func (c *Crawler) sendFileURLsToChannel(getFileUrls func(string, int) []string, blogName, dstDir string) {
	fetchNum := c.opt.Offset
	for fetchNum <= c.opt.PostNumPerBlog+c.opt.Offset {
		blogDstDir := path.Join(dstDir, blogName)
		fileUrls := getFileUrls(blogName, fetchNum)

		log.Printf("%d URLs are found on %s %d-%d / %d",
			len(fileUrls), blogName, fetchNum, fetchNum+20, c.opt.PostNumPerBlog+c.opt.Offset)

		fileUrls, err := filterExistFileUrls(fileUrls, blogDstDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, fileUrl := range fileUrls {
			c.resourceChannel <- kniv.Resource{
				Url:          fileUrl,
				ResourceType: "tumblr",
				DstPath:      blogDstDir,
			}
		}

		time.Sleep(c.opt.APIIntervalMilliSec)

		if len(fileUrls) == 0 {
			return
		}
		fetchNum += 20
	}
}
