package unicorn

import (
	"encoding/json"
	"fmt"
	"github.com/godoji/candlestick"
	"log"
	"marlin/internal/config"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetInfo() *candlestick.ExchangeInfo {

	log.Println("fetch binance spot exchange info")

	result := &candlestick.ExchangeInfo{
		Name:       "USA Stocks",
		ExchangeId: "US",
		BrokerId:   "UNICORN",
		LastUpdate: time.Now().UTC().Unix(),
		Symbols:    make(map[string]*candlestick.AssetInfo),
		Resolution: []int64{candlestick.Interval1d},
	}

	symbols := config.SymbolList(config.SourceUnicorn)
	for symbol := range symbols {
		info := &candlestick.AssetInfo{
			Identifier:         candlestick.NewAssetIdentifier(result.BrokerId, result.ExchangeId, symbol),
			Symbol:             symbol,
			Pair:               symbol + "USD",
			BaseAsset:          symbol,
			BaseAssetPrecision: 2,
			QuoteAsset:         "USD",
			QuotePrecision:     2,
			Constraints:        candlestick.TradeConstraints{},
			OnBoardDate:        math.MinInt64,
		}
		info.Splits = GetSplits(info.Identifier)
		result.Symbols[info.Identifier.ToString()] = info
	}

	return result
}

type SplitResponse struct {
	Date  string `json:"date"`
	Split string `json:"split"`
}

func GetSplits(target candlestick.AssetIdentifier) []candlestick.AssetSplit {

	url := fmt.Sprintf("%s/splits/%s.%s?api_token=%s&fmt=json", config.UnicornAPI, target.Symbol, target.Exchange, config.ServiceConfig().UnicornKey())
	req, err := http.Get(url)
	if err != nil {
		log.Println("could not retrieve split info")
		log.Fatalln(err)
	}
	if req.StatusCode != http.StatusOK {
		log.Println("could not retrieve split info")
		log.Fatalln(req.Status)
	}

	defer req.Body.Close()

	payload := make([]SplitResponse, 0)

	err = json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		log.Println("failed to decode splits")
		log.Fatalln(err)
	}

	results := make([]candlestick.AssetSplit, 0)
	for _, entry := range payload {
		ts, err := time.Parse("2006-01-02", entry.Date)
		if err != nil {
			log.Fatalln("could not parse split date")
		}

		splitParts := strings.Split(entry.Split, "/")
		n, err := strconv.ParseFloat(splitParts[0], 64)
		if err != nil {
			log.Fatalln("could not decode split ratio float")
		}
		d, err := strconv.ParseFloat(splitParts[1], 64)
		if err != nil {
			log.Fatalln("could not decode split ratio float")
		}

		results = append(results, candlestick.AssetSplit{
			Time:  ts.UTC().Unix(),
			Ratio: n / d,
		})
	}

	return results
}
