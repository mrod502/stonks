package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/next-alpha/stonks/stonks"
	"github.com/next-alpha/utils/logger"
)

var (
	servePort  = ":1901"
	server     *mux.Router
	dataDir    string
	symbolsDir string
	stonksDir  string

	allSymbols map[string]stonks.StockSymbol
	data       map[string][]stonks.Sentiment

	refreshInterval = 3 * time.Minute
	db              *stonks.MemCache
)

func init() {
	//path
	homeDir, _ := os.UserHomeDir()
	stonksDir = path.Join(homeDir, "stonks")
	symbolsDir = path.Join(stonksDir, "symbols")
	dataDir = path.Join(stonksDir, "data")
	stonks.DoStonks()

	//router
	server = mux.NewRouter()
	//server.HandleFunc("/sentiment/{symbol}/{begin-timestamp}/{end-timestamp}", grabSentiment)
	server.HandleFunc("/sentiment/{symbol}", allSymbolSentiment)
	server.HandleFunc("/", home)
	server.HandleFunc("/sentiment", allSentiment)
	server.HandleFunc("/feed/{n}", allSentimentSlice)
	server.HandleFunc("/trades", func(w http.ResponseWriter, r *http.Request) {
		var b []byte
		enableCORS(&w)
		res, err := http.Get("http://ai.rtsignal.io:1337/trades/IB")
		if err != nil {
			fmt.Printf("Request failed: %v", err)
			w.Write([]byte("[]"))
			return
		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			w.Write([]byte("[]"))
			return
		}

		res.Body.Close()
		r.Body.Close()
		w.Write(b)
	})
	server.HandleFunc("/quotes/{symbol}/{timestamp}", getQuotes)
	server.HandleFunc("/robots.txt", func(h http.ResponseWriter, r *http.Request) { h.Write([]byte("User-agent: *\nfuck off\n")) })
	server.HandleFunc("/archive/sentiment/{date}/{symbol}", getByDateSymbol)

	// data
	db = stonks.NewMemCache()
	allSymbols = make(map[string]stonks.StockSymbol)

	_, err := os.Stat(path.Join(symbolsDir, "symbols_all.json"))
	if !os.IsNotExist(err) {
		x := make([]stonks.StockSymbol, 0, 128)
		b, _ := ioutil.ReadFile(path.Join(symbolsDir, "symbols_all.json"))

		_ = json.Unmarshal(b, &x)
		for _, v := range x {
			allSymbols[v.ACTSymbol] = v
		}
	}

}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] != "" {
			servePort = os.Args[1]
		}
	}

	var ch = make(chan stonks.Sentiment, 256)

	go func() {
		for {
			err := stonks.LoadAllSentiment(ch)
			if err == nil {
				for {
					v := <-ch
					if _, ok := allSymbols[v.Symbol]; ok || (len(v.Symbol) < 5 && len(v.Symbol) > 2) {
						db.Set(v)
					}
				}
			}
			time.Sleep(refreshInterval)
		}
	}()

	go func() {
		for {
			err := http.ListenAndServe(servePort, server)
			if err != nil {
				logger.Error("server: ", err.Error())
				if strings.Contains(err.Error(), "missing port in address") {
					return
				}
				time.Sleep(time.Second)
			}
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)

	fmt.Println(fmt.Sprintf("ctrl+c pressed, exiting: %s", <-c))
	err := db.Sync()
	if err != nil {
		logger.Error("sync", err.Error())
	}
}

func allSymbolSentiment(h http.ResponseWriter, r *http.Request) {
	enableCORS(&h)
	vars := mux.Vars(r)
	logger.Info("sentiment: ", fmt.Sprintf("%10s", vars["symbol"]), " request origin ", r.RemoteAddr)
	symbol, ok := vars["symbol"]
	if !ok {
		h.Write([]byte{0})
		return
	}
	b, _ := json.Marshal(db.GetBySymbol(symbol))
	h.Write(b)

	return
}

func home(h http.ResponseWriter, r *http.Request) {
	enableCORS(&h)
	logger.Info("sentiment: home", " | request origin: ", r.RemoteAddr)
	h.Write([]byte("welcome to the salty spitoon, how tough are ya?"))
	return
}

func grabSentiment(h http.ResponseWriter, r *http.Request) {
	enableCORS(&h)
	vars := mux.Vars(r)

	symbol := vars["symbol"]
	logger.Info("sentiment: ", fmt.Sprintf("%10s", vars["symbol"]), fmt.Sprintf("%13s", vars["begin-timestamp"]), fmt.Sprintf("%13s", vars["end-timestamp"])+" | request origin: "+r.RemoteAddr)

	startTime, _ := strconv.ParseInt(vars["begin-timestamp"], 10, 64)
	endTime, _ := strconv.ParseInt(vars["end-timestamp"], 10, 64)

	b, _ := json.Marshal(db.GetSymbolByTime(symbol, startTime, endTime))

	h.Write(b)

}
func allSentiment(h http.ResponseWriter, r *http.Request) {
	enableCORS(&h)
	all := db.GetAll()
	logger.Info("sentiment", "all sentiment", "request origin", r.RemoteAddr, fmt.Sprintf("len: %d", len(all)))

	b, _ := json.Marshal(all)
	h.Write(b)
}
func allSentimentSlice(h http.ResponseWriter, r *http.Request) {
	enableCORS(&h)
	n, err := strconv.ParseInt(mux.Vars(r)["n"], 10, 64)
	if err != nil {
		n = 100
	}
	all := db.GetAllSlice()
	logger.Info("sentiment", "all sentiment", "request origin", r.RemoteAddr, fmt.Sprintf("len: %d", len(all)))
	if len(all) > 0 {
		b, _ := json.Marshal(all[max(len(all)-int(n), 0):])

		h.Write(b)
	}
	h.Write([]byte("[]"))
}

func getByDateSymbol(h http.ResponseWriter, r *http.Request) {
	enableCORS(&h)
	vars := mux.Vars(r)

	dateString := vars["date"]
	symbol := vars["symbol"]

	if dateString == "~" && symbol == "~" {
		s, err := stonks.SearchBuckets("")
		if err != nil {
			logger.Error("search", err.Error())
			h.Write([]byte{})
			return
		}
		b, err := json.Marshal(s)

		if err != nil {
			logger.Error("marshal", err.Error())
			h.Write([]byte{})
			return
		}
		_, err = h.Write(b)
		if err != nil {
			logger.Error("send", err.Error())
		}
		return
	}

	if symbol == "~" {
		s, err := stonks.SearchBuckets(dateString)
		if err != nil {
			logger.Error("search", err.Error())
			h.Write([]byte{})
			return
		}
		b, err := json.Marshal(s)

		if err != nil {
			logger.Error("marshal", err.Error())
			h.Write([]byte{})
			return
		}
		_, err = h.Write(b)
		if err != nil {
			logger.Error("send", err.Error())
		}
		return
	}

	if dateString == "~" {
		s, err := stonks.SearchBuckets(symbol)
		if err != nil {
			logger.Error("search", err.Error())
			h.Write([]byte{})
			return
		}
		b, err := json.Marshal(s)

		if err != nil {
			logger.Error("marshal", err.Error())
			h.Write([]byte{})
			return
		}
		_, err = h.Write(b)
		if err != nil {
			logger.Error("send", err.Error())
		}
		return
	}

	//get bucket
	sent, err := stonks.GetBucket(dateString + "$" + symbol)
	if err != nil {
		logger.Error("getbucket", err.Error())
		h.Write([]byte{})
		return
	}
	b, err := json.Marshal(sent)
	if err != nil {
		logger.Error("marshal", err.Error())
		h.Write([]byte{})
		return
	}
	h.Write(b)
	return

}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("access-control-allow-origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS,POST")

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getQuotes(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Connection", "close")

	res, err := http.Get("http://ai.rtsignal.io:1337" + r.RequestURI)
	if err != nil {
		w.Write([]byte("request failed"))

		return
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error("get", err.Error())
		w.Write([]byte("request failed"))
		return
	}
	w.Write(b)
	res.Body.Close()

}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
