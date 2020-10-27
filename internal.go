package stonks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cdipaolo/sentiment"
	"github.com/next-alpha/utils/logger"

	"github.com/turnage/graw/reddit"
)

//Sentiment - struct to represent sentiment of a reddit object
type Sentiment struct {
	Symbol      string  `json:"symbol"`
	Score       float64 `json:"score"` //0 (bad)-100 (good)
	Clout       float64 `json:"clout"` //small number -> not important, big number -> very important (upvotes minus downvotes)
	PostID      string  `json:"post_id"`
	CommentText string  `json:"comment_text"`
	Timestamp   int64   `json:"timestamp"`
}

//LoadAllSentiment -- read reddit
func LoadAllSentiment(sent chan Sentiment) (err error) {
	var allPosts []*reddit.Post
	bot, err := reddit.NewBotFromAgentFile(agentPath, 5*time.Second)
	if err != nil {
		return err
	}
	var ch = make(chan *reddit.Post, 256) //communicate between 2 goroutines

	// ---------------------------get top-level posts ----------------------------
	go func() {
		for {

			harvest, err := bot.Listing(testURL, "")
			if err != nil {
				logger.Error("harvest", err.Error())
				return
			}
			allPosts = harvest.Posts
			for _, v := range allPosts {
				ch <- v
			}
			time.Sleep(2 * time.Minute)
		}
	}()
	// ----------------------------------------------------------------------------
	// ------------------ get threads of each top-level post ----------------------
	go func() {

		var URL string
		for {
			p := <-ch
			URL = "https://www.reddit.com/r/wallstreetbets/comments/" +
				p.ID + ".json?sort=newest" //get top daily comments

			//get sentiment of top level post
			ss, _ := getPostSentiment(p)
			for _, v := range ss {
				sent <- v
			}

			if URL == "" {
				logger.Error("no url", URL)
				return
			}
			b, h, err := BrowserRequest(URL)
			if bytes.Contains(b, []byte("Too Many Requests")) {
				logger.Error("get", "too many requests, chilling for a bit")
				time.Sleep(2 * time.Minute)
			}
			if err != nil {
				logger.Error("request", err.Error())
				time.Sleep(1 * time.Minute)
				continue
			}
			remainingCalls, reset := checkRemainingCalls(h)

			if remainingCalls == 0 {
				if reset != 0 {
					logger.Warn("Calls", fmt.Sprintf("Remaining:%v | Reset: %v", remainingCalls, reset))
					time.Sleep(time.Duration(reset+1) * time.Second)
				} else {
					time.Sleep(10 * time.Second)
				}
			} else {
				time.Sleep(500*time.Millisecond + (time.Second * (1 + time.Duration(reset/remainingCalls))))
			}

			var res JSONResponse
			err = json.Unmarshal(b, &res)
			if err != nil {
				logger.Error("Unmarshal", err.Error())
			}
			if len(res) < 2 {
				logger.Error("Res", "No response")
				continue
			}
			posts := res[1].GetAllComments()
			logger.Info("posts", fmt.Sprintf("num comments: %d", len(posts)))

			for _, p := range posts {
				s, _ := getCommentSentiment(p)
				for _, v := range s {
					if v.Symbol != "" {
						sent <- v
					}
				}
			}
		}

	}()
	// -----------------------------------------------------------------------------
	return
}

func getParagraphSentiment(sentences []string) float64 {
	if len(sentences) == 0 {
		return 0
	}

	var feels float64
	var sentenceCount = float64(len(sentences))
	var sent *sentiment.Analysis
	if sentenceCount < 1 {
		return 50 //return neutral sentiment if no sentences
	}

	for _, sentence := range sentences {
		sent = model.SentimentAnalysis(sentence, sentiment.English)

		feels += 100 * float64(sent.Score)
	}

	return feels / sentenceCount

}

func pullHyperlinks(text string) (notlinks string, links []string) {
	//pattern for hyperlinks

	linksSub := patternHyperlink.FindAllStringSubmatch(text, -1)

	for _, link := range linksSub {
		if len(link) == 2 {
			links = append(links, link[1])
			//	writeLog(link[1])
			text = strings.Replace(text, link[0], " ", -1)
		}
	}
	notlinks = text

	return notlinks, links
}

//get the sentences from a post or whatever
func postSentences(post string) []string {
	post = strings.Replace(post, "&#x200B;", "", -1)
	return strings.FieldsFunc(post, func(r rune) bool {
		if r == '.' || r == '\n' || r == ':' {
			return true
		}
		return false
	})
}

func getCommentSentiment(p Comment) (sent []Sentiment, links []string) {

	//text, links := pullHyperlinks(p.SelfText)
	text := p.Data.Body
	if strings.Contains(text, "I am a bot") {
		return
	}
	sentences := postSentences(text)

	feel := getParagraphSentiment(sentences)
	symbols := getSymbols(text)

	clout := p.Ups - p.Downs

	for _, v := range symbols {

		s := Sentiment{
			Score:       feel,
			Clout:       float64(clout),
			Symbol:      v,
			Timestamp:   int64(p.CreatedUTC),
			PostID:      p.ID,
			CommentText: p.Data.Body,
		}

		sent = append(sent, s)
	}

	return
}
func getPostSentiment(p *reddit.Post) (sent []Sentiment, links []string) {

	text, links := pullHyperlinks(p.SelfText)

	sentences := postSentences(text)

	feel := getParagraphSentiment(sentences)
	symbols := getSymbols(text)

	clout := p.Ups - p.Downs

	for _, v := range symbols {
		if !ignored[v] && v != "" {
			s := Sentiment{Score: feel, Clout: float64(clout), Symbol: v, Timestamp: int64(p.CreatedUTC), PostID: p.ID}
			sent = append(sent, s)
		}
	}

	return
}

func getSymbols(text string) (symbols []string) {

	var matchMap = make(map[string]bool)

	matches := patternSymbols.FindAllString(text, -1)

	for _, match := range matches {

		if !ignored[match] {
			matchMap[match] = true
		}

	}
	matches2 := patternSlash.FindAllStringSubmatch(text, -1)

	for _, match := range matches2 {
		if len(match) == 2 {
			if !ignored[match[1]] {
				matchMap[match[1]] = true
			}
		}
	}

	matches2 = patternDollarSign.FindAllStringSubmatch(text, -1)

	for _, match := range matches2 {
		if len(match) == 2 {
			if !ignored[match[1]] {
				matchMap[match[1]] = true
			}
		}
	}
	symbols = make([]string, 0, len(matchMap))
	for v := range matchMap {
		symbols = append(symbols, v)
	}

	return
}

func getSubmatch(a []string) string {
	if len(a) == 2 {
		return a[1]
	}
	return ""
}

func processMore(m MoreData, comment CommentData) (c []CommentData, err error) {

	for _, v := range m.Children {
		time.Sleep(5 * time.Second) // don't get banned
		var jr JSONResponse
		b, _, _ := BrowserRequest(WSBURL + comment.LinkID + "/" + comment.LinkTitle + "/" + v)
		err = json.Unmarshal(b, &jr)

		if err != nil {
			continue
		}
		for _, v := range c {
			logger.Error("", v.Body)
		}
	}
	return
}

func checkRemainingCalls(header http.Header) (remaining, reset int64) {

	if v := header.Get("x-ratelimit-remaining"); v != "" {

		remaining, _ = strconv.ParseInt(v, 10, 64)
		reset, _ = strconv.ParseInt(header.Get("x-ratelimit-reset"), 10, 64)
	}
	return
}
