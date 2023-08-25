package binance

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/godoji/candlestick"
	"log"
	"marlin/internal/throw"
	"time"
)

func FetchCandles(from time.Time, target candlestick.AssetIdentifier) ([]candlestick.Candle, throw.Exception) {
	var err error
	var candles []candlestick.Candle
	switch target.Exchange {
	case "PERP":
		candles, err = fetchFuturesCandles(from, target.Symbol)
	case "SPOT":
		candles, err = fetchSpotCandles(from, target.Symbol)
	default:
		return nil, throw.ErrInvalidExchange
	}
	if err != nil {
		log.Printf("failed fetching block %s at %s: %s\n", target.Symbol, from.UTC().Format(time.RFC3339), err.Error())
		return nil, throw.New(err, throw.ErrKindUnexpected)
	}
	return candles, nil
}

func fetchFuturesCandles(from time.Time, symbol string) ([]candlestick.Candle, error) {

	// Fetch candles from Binance
	var klines []*futures.Kline

	if from.Unix() > time.Now().UTC().Unix()+60*15 {
		klines = make([]*futures.Kline, 0)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		historyService := futuresClient.NewKlinesService()
		var err error
		klines, err = historyService.
			Interval("1m").
			Symbol(symbol).
			Limit(fetchLimit).
			StartTime(from.UnixMilli()).
			Do(ctx)
		cancel()

		// Forward error if any
		if err != nil {
			return nil, err
		}
	}

	filtered := fillMissingCandles(futureToSpot(klines), from, symbol)

	candles := make([]candlestick.Candle, 1000)
	for i, k := range filtered {
		candles[i] = klineToCandle(
			k.Open,
			k.High,
			k.Low,
			k.Close,
			k.Volume,
			k.TradeNum,
			k.TakerBuyQuoteAssetVolume,
			k.OpenTime/1000,
		)
		c := candles[i]
		if c.Open == 0 && c.High == 0 && c.Low == 0 && c.Close == 0 {
			candles[i].Missing = true
		}
	}

	return candles, nil
}

func fetchSpotCandles(from time.Time, symbol string) ([]candlestick.Candle, error) {

	// Fetch candles from Binance
	var klines []*binance.Kline

	if from.Unix() > time.Now().UTC().Unix()+60*15 {
		klines = make([]*binance.Kline, 0)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		historyService := spotClient.NewKlinesService()
		var err error
		klines, err = historyService.
			Interval("1m").
			Symbol(symbol).
			Limit(fetchLimit).
			StartTime(from.UnixMilli()).
			Do(ctx)
		cancel()

		// Forward error if any
		if err != nil {
			return nil, err
		}
	}

	filtered := fillMissingCandles(klines, from, symbol)

	candles := make([]candlestick.Candle, 1000)
	for i, k := range filtered {
		candles[i] = klineToCandle(
			k.Open,
			k.High,
			k.Low,
			k.Close,
			k.Volume,
			k.TradeNum,
			k.TakerBuyQuoteAssetVolume,
			k.OpenTime/1000,
		)
		c := candles[i]
		if c.Open == 0 && c.High == 0 && c.Low == 0 && c.Close == 0 {
			candles[i].Missing = true
		}
	}

	return candles, nil
}
