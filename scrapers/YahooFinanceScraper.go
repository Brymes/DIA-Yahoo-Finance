package scrapers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type YahooFinanceScraper struct {
	Close        chan bool
	PriceChannel chan float64
	// error handling; to read error or closed, first acquire read lock
	// only cleanup method should hold write lock
	ticker *time.Ticker
	pair   string
}

type YahooFinanceScraperResponse struct {
	Date     string
	Open     float64
	High     float64
	Low      float64
	Close    float64
	AdjClose float64
	Volume   float64
}

// NewYahooFinanceScraper returns a new YahooFinanceScraper initialized with default values.
// The instance is asynchronously scraping as soon as it is created.
func NewYahooFinanceScraper(pair string, scrape bool, refreshDelay int) *YahooFinanceScraper {
	delay := time.Duration(refreshDelay) * time.Second
	s := &YahooFinanceScraper{
		Close:        make(chan bool),
		PriceChannel: make(chan float64),
		ticker:       time.NewTicker(delay),
		pair:         pair,
	}
	if scrape {
		go s.mainLoop()
	}
	return s
}

func (s *YahooFinanceScraper) mainLoop() {
	s.Update()
	for {
		select {
		case <-s.ticker.C:
			s.Update()
		case <-s.Close: // user requested shutdown
			close(s.PriceChannel)
			log.Printf("YahooFinanceScraper shutting down")
			return
		}
	}
}

func (s *YahooFinanceScraper) ParseUrl() string {
	//https://query2.finance.yahoo.com/v7/finance/download/EURUSD=X
	baseUrl := "https://query2.finance.yahoo.com/v7/finance/download/%s=X"

	return fmt.Sprintf(baseUrl, s.pair)
}

func (s *YahooFinanceScraper) Update() {
	url := s.ParseUrl()

	method := "GET"
	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := csv.NewReader(bytes.NewReader(body))
	records, _ := reader.ReadAll()

	var data []YahooFinanceScraperResponse
	for _, record := range records[1:] {
		open, err := strconv.ParseFloat(record[1], 64)
		high, err := strconv.ParseFloat(record[2], 64)
		low, err := strconv.ParseFloat(record[3], 64)
		closing, err := strconv.ParseFloat(record[4], 64)
		adjClose, err := strconv.ParseFloat(record[5], 64)
		volume, err := strconv.ParseFloat(record[6], 64)

		if err != nil {
			log.Println(err)
			panic("Invalid Response from Server")
		}
		resp := YahooFinanceScraperResponse{
			Date:     record[0],
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closing,
			AdjClose: adjClose,
			Volume:   volume,
		}
		data = append(data, resp)
	}

	s.PriceChannel <- data[0].AdjClose
}
