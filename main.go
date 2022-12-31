package main

import (
	"DIA-Yahoo-Finance-Scraper/scrapers"
	"log"
)

func main() {
	x := scrapers.NewYahooFinanceScraper("EURUSD", true, 10)
	log.Println(x)

	c := <-x.PriceChannel
	log.Println(c)

	x.Close <- true

	log.Println("done")
}
