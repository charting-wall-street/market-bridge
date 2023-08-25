package binance

import (
	"context"
	"errors"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/godoji/candlestick"
	"log"
	"marlin/internal/config"
	"math"
	"strconv"
	"sync"
	"time"
)

func isFutureSymbolValid(s futures.Symbol) bool {
	cryptoPairs := config.SymbolList(config.SourceBinance)
	if _, ok := cryptoPairs[s.BaseAsset]; !ok {
		return false
	}
	if s.QuoteAsset != "USDT" {
		return false
	}
	if s.ContractType != "PERPETUAL" {
		return false
	}
	if s.Status != "TRADING" {
		return false
	}
	return true
}

func isSpotSymbolValid(s binance.Symbol) bool {
	cryptoPairs := config.SymbolList(config.SourceBinance)
	if _, ok := cryptoPairs[s.BaseAsset]; !ok {
		return false
	}
	if !s.IsSpotTradingAllowed {
		return false
	}
	if !s.IsMarginTradingAllowed {
		return false
	}
	if s.QuoteAsset != "USDT" {
		return false
	}
	if s.Status != "TRADING" {
		return false
	}
	return true
}

func fetchFuturesExchangeInfo() *futures.ExchangeInfo {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	info, err := futuresClient.NewExchangeInfoService().Do(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	return info
}

func fetchSpotExchangeInfo() *binance.ExchangeInfo {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	info, err := spotClient.NewExchangeInfoService().Do(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	return info
}

func GetFuturesInfo() *candlestick.ExchangeInfo {

	log.Println("fetch binance futures exchange info")

	result := &candlestick.ExchangeInfo{
		Name:       "Futures Trading",
		ExchangeId: "PERP",
		BrokerId:   "BINANCE",
		LastUpdate: time.Now().UTC().Unix(),
		Symbols:    make(map[string]*candlestick.AssetInfo),
		Resolution: []int64{candlestick.Interval1m},
	}
	info := fetchFuturesExchangeInfo()
	onBoardDateMap := make(map[string]int64)
	onBoardLock := sync.Mutex{}
	count := 0

	var wg sync.WaitGroup
	sem := make(chan struct{}, 20)

	for _, s := range info.Symbols {
		if !isFutureSymbolValid(s) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(symbol string) {
			defer wg.Done()
			onBoard, err := getFuturesOnBoardDate(symbol)
			if err != nil {
				log.Fatal(err)
			}
			onBoardLock.Lock()
			onBoardDateMap[symbol] = onBoard
			onBoardLock.Unlock()
			<-sem
		}(s.Symbol)
	}

	wg.Wait()

	var err error
	for _, s := range info.Symbols {
		if !isFutureSymbolValid(s) {
			continue
		}

		minPrice := 0.0
		maxPrice := math.MaxFloat64
		tickSize := 0.000001
		maxQuantity := 10000000.0
		minQuantity := 0.001
		stepSize := 0.001
		maxNumOrders := 100
		minNotional := 5.0
		for _, filter := range s.Filters {
			switch filter["filterType"] {
			case "PRICE_FILTER":
				maxPrice, err = strconv.ParseFloat(filter["maxPrice"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				minPrice, err = strconv.ParseFloat(filter["minPrice"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				tickSize, err = strconv.ParseFloat(filter["tickSize"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
			case "MARKET_LOT_SIZE":
				maxQuantity, err = strconv.ParseFloat(filter["maxQty"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				minQuantity, err = strconv.ParseFloat(filter["minQty"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				stepSize, err = strconv.ParseFloat(filter["stepSize"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
			case "MAX_NUM_ORDERS":
				maxNumOrders = int(filter["limit"].(float64))
			case "MIN_NOTIONAL":
				minNotional, err = strconv.ParseFloat(filter["notional"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		identifier := candlestick.NewAssetIdentifier(result.BrokerId, result.ExchangeId, s.Symbol)
		symbolInfo := &candlestick.AssetInfo{
			Identifier:         identifier,
			Symbol:             identifier.ToString(),
			Pair:               s.Symbol,
			BaseAssetPrecision: s.BaseAssetPrecision,
			BaseAsset:          s.BaseAsset,
			QuotePrecision:     s.QuotePrecision,
			QuoteAsset:         s.QuoteAsset,
			OnBoardDate:        onBoardDateMap[s.Symbol],
			Splits:             []candlestick.AssetSplit{},
			Constraints: candlestick.TradeConstraints{
				MaxPrice:     maxPrice,
				MinPrice:     minPrice,
				TickSize:     tickSize,
				MaxQuantity:  maxQuantity,
				MinQuantity:  minQuantity,
				StepSize:     stepSize,
				MaxNumOrders: maxNumOrders,
				MinNotional:  minNotional,
			},
		}

		result.Symbols[identifier.ToString()] = symbolInfo
		count++

	}

	return result
}

func GetSpotInfo() *candlestick.ExchangeInfo {

	log.Println("fetch binance spot exchange info")

	result := &candlestick.ExchangeInfo{
		Name:       "Spot Trading",
		ExchangeId: "SPOT",
		BrokerId:   "BINANCE",
		LastUpdate: time.Now().UTC().Unix(),
		Symbols:    make(map[string]*candlestick.AssetInfo),
		Resolution: []int64{candlestick.Interval1m},
	}
	info := fetchSpotExchangeInfo()
	onBoardDateMap := make(map[string]int64)
	onBoardLock := sync.Mutex{}
	count := 0

	var wg sync.WaitGroup
	sem := make(chan struct{}, 20)
	for _, s := range info.Symbols {
		if !isSpotSymbolValid(s) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(symbol string) {
			defer wg.Done()
			onBoard, err := getSpotOnBoardDate(symbol)
			if err != nil {
				log.Fatal(err)
			}
			onBoardLock.Lock()
			onBoardDateMap[symbol] = onBoard
			onBoardLock.Unlock()
			<-sem
		}(s.Symbol)
	}
	wg.Wait()

	var err error
	for _, s := range info.Symbols {

		if !isSpotSymbolValid(s) {
			continue
		}

		minPrice := 0.0
		maxPrice := math.MaxFloat64
		tickSize := 0.000001
		maxQuantity := 10000000.0
		minQuantity := 0.001
		stepSize := 0.001
		maxNumOrders := 100
		minNotional := 5.0
		for _, filter := range s.Filters {
			switch filter["filterType"] {
			case "PRICE_FILTER":
				maxPrice, err = strconv.ParseFloat(filter["maxPrice"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				minPrice, err = strconv.ParseFloat(filter["minPrice"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				tickSize, err = strconv.ParseFloat(filter["tickSize"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
			case "MARKET_LOT_SIZE":
				maxQuantity, err = strconv.ParseFloat(filter["maxQty"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				minQuantity, err = strconv.ParseFloat(filter["minQty"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
				stepSize, err = strconv.ParseFloat(filter["stepSize"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
			case "MAX_NUM_ORDERS":
				maxNumOrders = int(filter["maxNumOrders"].(float64))
			case "MIN_NOTIONAL":
				minNotional, err = strconv.ParseFloat(filter["minNotional"].(string), 64)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		identifier := candlestick.NewAssetIdentifier(result.BrokerId, result.ExchangeId, s.Symbol)
		symbolInfo := &candlestick.AssetInfo{
			Identifier:         identifier,
			Symbol:             identifier.ToString(),
			Pair:               s.Symbol,
			BaseAssetPrecision: s.BaseAssetPrecision,
			QuotePrecision:     s.QuotePrecision,
			QuoteAsset:         s.QuoteAsset,
			BaseAsset:          s.BaseAsset,
			OnBoardDate:        onBoardDateMap[s.Symbol],
			Splits:             []candlestick.AssetSplit{},
			Constraints: candlestick.TradeConstraints{
				MaxPrice:     maxPrice,
				MinPrice:     minPrice,
				TickSize:     tickSize,
				MaxQuantity:  maxQuantity,
				MinQuantity:  minQuantity,
				StepSize:     stepSize,
				MaxNumOrders: maxNumOrders,
				MinNotional:  minNotional,
			},
		}

		result.Symbols[symbolInfo.Identifier.ToString()] = symbolInfo
		count++

	}

	return result
}

func getFuturesOnBoardDate(symbol string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	historyService := futuresClient.NewKlinesService()
	klines, err := historyService.
		Interval("1m").
		Symbol(symbol).
		Limit(1).
		StartTime(0).
		Do(ctx)
	cancel()
	if err != nil {
		return 0, err
	}
	if len(klines) == 0 {
		log.Printf("no candles returned when requesting first candle of %s\n", symbol)
		return 0, errors.New("no candles received")
	}
	return klines[0].OpenTime / 1000, nil
}

func getSpotOnBoardDate(symbol string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	historyService := spotClient.NewKlinesService()
	klines, err := historyService.
		Interval("1m").
		Symbol(symbol).
		Limit(1).
		StartTime(0).
		Do(ctx)
	cancel()
	if err != nil {
		return 0, err
	}
	if len(klines) == 0 {
		log.Printf("no candles returned when requesting first candle of %s\n", symbol)
		return 0, errors.New("no candles received")
	}
	return klines[0].OpenTime / 1000, nil
}
