package arbiter

import (
	"encoding/json"
	"github.com/godoji/candlestick"
	"log"
	"marlin/internal/binance"
	"marlin/internal/config"
	"marlin/internal/throw"
	"marlin/internal/unicorn"
	"os"
	"sync"
	"time"
)

var exchangeInfoCache *candlestick.ExchangeList = nil
var exchangeInfoLock = sync.Mutex{}
var exchangeIsFetching = false

func FetchHistorical(target candlestick.AssetIdentifier, from int64, interval int64) ([]candlestick.Candle, throw.Exception) {
	tsFrom := time.Unix(from, 0).UTC()
	switch target.Broker {
	case config.SourceUnicorn:
		switch interval {
		case candlestick.Interval1d:
			return unicorn.FetchHistorical(target)
		default:
			return nil, throw.ErrIntervalNotSupported
		}
	case config.SourceBinance:
		if from == 0 {
			return nil, throw.ErrInvalidFromParameter
		}
		switch interval {
		case candlestick.Interval1m:
			return binance.FetchCandles(tsFrom, target)
		default:
			return nil, throw.ErrIntervalNotSupported
		}
	default:
		return nil, throw.ErrInvalidSource
	}
}

func FetchLatest(target candlestick.AssetIdentifier, from int64) ([]candlestick.Candle, throw.Exception) {
	switch target.Broker {
	case config.SourceUnicorn:
		return nil, throw.ErrSourceNotSupported
	case config.SourceBinance:
		return binance.FetchLatest(from, target)
	default:
		return nil, throw.ErrInvalidSource
	}
}

func refreshExchangeInfo() {
	result := &candlestick.ExchangeList{
		Exchanges: make([]*candlestick.ExchangeInfo, 0),
		BrokerInfo: map[string]*candlestick.BrokerInfo{
			"BINANCE": {Name: "Binance"},
			"UNICORN": {Name: "Unicorn"},
		},
	}

	result.Exchanges = append(result.Exchanges, unicorn.GetInfo())
	result.Exchanges = append(result.Exchanges, binance.GetSpotInfo())
	result.Exchanges = append(result.Exchanges, binance.GetFuturesInfo())

	exchangeInfoLock.Lock()
	defer exchangeInfoLock.Unlock()
	exchangeInfoCache = result
	writeInfoToDisk()
	exchangeIsFetching = false
}

func ExchangeInfo() *candlestick.ExchangeList {

	exchangeInfoLock.Lock()

	// Load from disk if not data has been retrieved
	if exchangeInfoCache == nil {
		loadInfoFromDisk()
	}

	if config.ServiceConfig().IsOffline() {
		if exchangeInfoCache == nil {
			log.Fatalln("no cached exchange info found")
		}

		defer exchangeInfoLock.Unlock()
		return exchangeInfoCache
	}

	// Check if we have any data
	if exchangeInfoCache != nil {
		now := time.Now().UTC().Unix()
		isUpToDate := true
		for _, exchange := range exchangeInfoCache.Exchanges {
			if (now - exchange.LastUpdate) > 8*60*60 {
				isUpToDate = false
				break
			}
		}
		// Update the exchange data in the background
		if !isUpToDate && !exchangeIsFetching {
			exchangeIsFetching = true
			go refreshExchangeInfo()
		}
		defer exchangeInfoLock.Unlock()
		return exchangeInfoCache
	}

	// Fetch data synchronously
	exchangeInfoLock.Unlock()
	refreshExchangeInfo()

	// Return exchange data
	exchangeInfoLock.Lock()
	defer exchangeInfoLock.Unlock()
	return exchangeInfoCache
}

func writeInfoToDisk() {
	file, err := os.Create("./data/exchange.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewEncoder(file).Encode(exchangeInfoCache)
	if err != nil {
		log.Fatal(err)
	}
	_ = file.Close()
}

func loadInfoFromDisk() bool {
	path := "./data/exchange.json"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	e := new(candlestick.ExchangeList)
	err = json.NewDecoder(file).Decode(e)
	if err != nil {
		log.Fatal(err)
	}
	_ = file.Close()
	log.Println("existing exchange info for binance found")
	exchangeInfoCache = e
	return true
}
