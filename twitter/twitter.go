package twitter

import (
	"github.com/mpppk/kniv/kniv"
)

func init() {
	kniv.RegisterCrawlerFactory(&CrawlerFactory{})
}
