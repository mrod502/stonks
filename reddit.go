package stonks

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/next-alpha/utils/logger"
	"github.com/sridharv/reddit-go"
)

//GetStream -- get a reddit stream
func GetStream(configFile string) (links []*reddit.Link) {
	logFile, err := os.OpenFile("stream.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("logger", "error opening log file:", err.Error())
		return
	}
	defer logFile.Close()
	//	fileLogger = log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)

	cfg, err = reddit.LoadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := cfg.AuthScript(http.DefaultClient); err != nil {
		log.Fatal(err)
	}

	// Print top posts of the day from /r/golang
	stream := cfg.Stream(http.DefaultClient, &reddit.TopPosts{SubReddit: "wallstreetbets", Duration: reddit.TopDay})

	for stream.Next() {
		thing := stream.Thing()

		link, ok := thing.Data.(*reddit.Link)

		if ok {
			links = append(links, link)
		}
		if strings.Contains(link.URL, "discussion") {
			//getComment(getCommentId(link.URL), "wallstreetbets")

			break
			//configGet()
		}
		symbols := getSymbols(link.Title)
		str := ""
		for _, v := range symbols {
			str += v + ", "
		}
		logger.Info(link.URL, str)

	}
	if err := stream.Error(); err != nil {
		logger.Error("stream", err.Error())
	}

	return
}

//BrowserRequest -- pretend to be a browser so we can get comments
func BrowserRequest(url string, authority ...string) (b []byte, rh http.Header, err error) {

	r, _ := http.NewRequest("GET", url, nil)

	r.Header.Set("upgrade-insecure-requests", "1")
	r.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36")
	r.Header.Set("accept-language", "en-US,en;q=0.9,hr-HR;q=0.8,hr;q=0.7,ru-RU;q=0.6,ru;q=0.5")
	r.Header.Set("scheme", "https")
	r.Header.Set("authority", "www.reddit.com")
	if len(authority) == 1 {
		r.Header.Set("authority", authority[0])
	}
	cli := http.DefaultClient

	resp, err := cli.Do(r)
	if err != nil {
		return
	}
	rh = resp.Header

	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return
}
