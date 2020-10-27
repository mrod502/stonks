package stonks

import (
	"errors"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/cdipaolo/sentiment"

	"github.com/sridharv/reddit-go"
)

//Reddit kinds
const (
	KListing   = "Listing"
	KComment   = "t1"
	KAccount   = "t2"
	KLink      = "t3"
	KMessage   = "t4"
	KSubreddit = "t5"
	KAward     = "t6"
	KMore      = "more"
	WSBURL     = "https://www.reddit.com/r/wallstreetbets/"
)

//Errors, private vars, etc
var (
	ErrNotLink    = errors.New("not a link datatype")
	ErrNotComment = errors.New("not a comment datatype")

	logFile    *os.File
	cfg        *reddit.Config
	configLock sync.Mutex

	patternTimestamp   = regexp.MustCompile(`"created_utc": ([0-9]*)`)
	dailyDiscussionURL string

	//all-caps things we ignore
	ignored = map[string]bool{
		"CEO":  true,
		"RSI":  true,
		"RIP":  true,
		"WSB":  true,
		"JPOW": true,
		"FED":  true,
		"AND":  true,
		"OR":   true,
		"FUT":  true,
		"PUT":  true,
		"SHT":  true,
		"CBOE": true,
		"NYSE": true,
		"SWAP": true,
		"DAY":  true,
		"DATE": true,
		"IPO":  true,
		"AKA":  true,
		"EOD":  true,
		"EOW":  true,
		"EOM":  true,
		"EOY":  true,
		"AUM":  true,
		"YOLO": true,
		"FD":   true,
		"CALL": true,
		"DAMN": true,
	}

	SyncInterval            = 90 * time.Second
	expireTime        int64 = 60 * 60 * 6
	timeZone          *time.Location
	home              string
	testURL           = "/r/wallstreetbets"
	analysis          *sentiment.Analysis
	agentPath         string
	model, _          = sentiment.Restore()
	patternSymbols    = regexp.MustCompile(`[A-Z]{2,10}`) // will lose any 1-char tickers
	patternDollarSign = regexp.MustCompile(`\$([A-Z]+)`)
	patternSlash      = regexp.MustCompile(`/([A-Z]+)`)
	patternHyperlink  = regexp.MustCompile(`\[[^\]]*\]\(([^\)]*)\)`)
	bear              = `\ud83c\udf08\ud83d\udc3b`
	mux               = sync.RWMutex{}
)
