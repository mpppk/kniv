package tumblr

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"
	"time"

	"os"

	"sync"

	"github.com/MariaTerzieva/gotumblr"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/mpppk/kniv/etc"
	"github.com/skratchdot/open-golang/open"
)

type VideoPost struct {
	gotumblr.BasePost
	VideoUrl string `json:"video_url"`
}

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
}

type Client struct {
	*gotumblr.TumblrRestClient
	opt *Opt
}

func (c *Client) GetPhotoUrls(blogName string, offset int) []string {
	apiOpt := map[string]string{"offset": fmt.Sprint(offset)}
	photoRes := c.Posts(blogName, "photo", apiOpt)
	return GetImageUrlsFromAPIResponse(ConvertJsonToPhotoPosts(photoRes.Posts))
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

func fetchURL(wg *sync.WaitGroup, q chan string, dstDir string, sleepMilliSec time.Duration) {
	defer wg.Done()
	queueSize := 0
	for {
		fileUrl, ok := <-q // closeされると ok が false になる
		if !ok {
			fmt.Println("url fetching is terminated")
			return
		}

		if len(q) != queueSize {
			queueSize = len(q)
			log.Printf("current URL queue size: %d\n", queueSize)
		}

		_, err := img.Download(fileUrl, dstDir)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(sleepMilliSec * time.Millisecond)
	}
}

func Crawl(opt *Opt) {
	photoDstDir := "photos"
	if v, ok := opt.DstDirMap["photo"]; ok {
		photoDstDir = v
	}
	//videoDstDir := "videos"
	//if v, ok := opt.DstDirMap["video"]; ok {
	//	videoDstDir = v
	//}

	maxBlogNum := opt.MaxBlogNum
	rawClient := gotumblr.NewTumblrRestClient(
		opt.ConsumerKey,
		opt.ConsumerSecret,
		opt.OauthToken,
		opt.OauthSecret,
		"callback_url",
		"http://api.tumblr.com",
	)

	client := Client{TumblrRestClient: rawClient, opt: opt}

	var wg sync.WaitGroup
	wg.Add(1)
	q := make(chan string, 100000)
	go fetchURL(&wg, q, photoDstDir, 3000)

	blogNames := client.GetBlogNames(maxBlogNum)
	requestCount := 0
	for i, blogName := range blogNames {
		fmt.Printf("---- fetch from %s %d/%d----\n", blogName, i, len(blogNames))
		client.fetchURLs(blogName, photoDstDir, requestCount, q)
		//client.fetchURLs(blogName, photoDstDir, requestCount, q)
		//for fetchNum <= postNumPerBlog+offset {
		//	photoUrls := client.GetPhotoUrls(blogName, fetchNum)
		//	requestCount++
		//	log.Printf("%d photo URLs are found on %s %d-%d / %d request: %d",
		//		len(photoUrls), blogName, fetchNum, fetchNum+20, postNumPerBlog+offset, requestCount)
		//	if len(photoUrls) == 0 {
		//		time.Sleep(apiInterval * time.Millisecond)
		//		break
		//	}
		//
		//	photoUrls, err := filterExistFileUrls(photoUrls, photoDstDir)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//
		//	if len(photoUrls) == 0 {
		//		time.Sleep(apiInterval * time.Millisecond)
		//		break
		//	}
		//
		//	for _, photoUrl := range photoUrls {
		//		q <- photoUrl
		//	}
		//
		//	fetchNum += 20
		//	time.Sleep(apiInterval * time.Millisecond)
		//}

		//fetchNum = offset
		//for fetchNum <= postNumPerBlog+offset {
		//	apiOpt := map[string]string{"offset": fmt.Sprint(fetchNum)}
		//	videoRes := rawClient.Posts(blogName, "video", apiOpt)
		//	requestCount++
		//
		//	videoUrls, err := GetVideoUrls(ConvertJsonToVideoPosts(videoRes.Posts))
		//	if err != nil {
		//		log.Print(err)
		//	}
		//
		//	log.Printf("%d video URLs are found on %s %d-%d / %d request: %d",
		//		len(videoUrls), blogName, fetchNum, fetchNum+20, postNumPerBlog+offset, requestCount)
		//
		//	if len(videoUrls) == 0 {
		//		time.Sleep(apiInterval * time.Millisecond)
		//		break
		//	}
		//
		//	downloadNum, err := img.DownloadFiles(videoUrls, path.Join(videoDstDir, blogName), downloadInterval)
		//	if err != nil {
		//		log.Print(err)
		//	}
		//
		//	if downloadNum == 0 {
		//		time.Sleep(apiInterval * time.Millisecond)
		//		break
		//	}
		//
		//	fetchNum += 20
		//	time.Sleep(apiInterval * time.Millisecond)
		//}
	}
}
func (c *Client) fetchURLs(blogName, dstDir string, requestCount int, q chan string) {
	fetchNum := c.opt.Offset
	for fetchNum <= c.opt.PostNumPerBlog+c.opt.Offset {
		requestCount++
		fileUrls := c.GetPhotoUrls(blogName, c.opt.Offset)

		log.Printf("%d photo URLs are found on %s %d-%d / %d request: %d",
			len(fileUrls), blogName, fetchNum, fetchNum+20, c.opt.PostNumPerBlog+c.opt.Offset, requestCount)

		if len(fileUrls) == 0 {
			time.Sleep(c.opt.APIIntervalMilliSec * time.Millisecond)
			break
		}

		fileUrls, err := filterExistFileUrls(fileUrls, dstDir)
		if err != nil {
			log.Fatal(err)
		}

		if len(fileUrls) == 0 {
			time.Sleep(c.opt.APIIntervalMilliSec * time.Millisecond)
			break
		}

		for _, photoUrl := range fileUrls {
			q <- photoUrl
		}

		fetchNum += 20
		time.Sleep(c.opt.APIIntervalMilliSec * time.Millisecond)
	}
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
	fileName, err := img.GetFileNameFromUrl(fileUrl)
	if err != nil {
		return false, err
	}

	if !img.IsExist(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return false, err
		}
	}

	if img.IsExist(path.Join(dir, fileName)) {
		return true, nil
	}
	return false, nil
}

func ConvertJsonToVideoPosts(jsonPosts []json.RawMessage) []VideoPost {
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

func ConvertJsonToPhotoPosts(jsonPosts []json.RawMessage) []gotumblr.PhotoPost {
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

func GetVideoUrls(videoPosts []VideoPost) ([]string, error) {
	var videoUrls []string
	for _, post := range videoPosts {
		if post.PostType != "video" {
			continue
		}
		videoUrls = append(videoUrls, post.VideoUrl)
	}
	return videoUrls, nil
}

func GetImageUrlsFromAPIResponse(photoPosts []gotumblr.PhotoPost) []string {
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
