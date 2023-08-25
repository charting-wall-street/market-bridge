package binance

import (
	"context"
	"github.com/godoji/candlestick"
	"log"
	"marlin/internal/throw"
	"time"
)

func FetchLatest(from int64, target candlestick.AssetIdentifier) ([]candlestick.Candle, throw.Exception) {

	var err error
	var candles []candlestick.Candle

	switch target.Exchange {
	case "PERP":
		candles, err = fetchFuturesLatest(from, target.Symbol)
	case "SPOT":
		candles, err = fetchSpotLatest(from, target.Symbol)
	default:
		log.Printf("invalid market type \"%s\"\n", target.Exchange)
		return nil, throw.ErrInvalidExchange
	}

	if err != nil {
		ts := time.Unix(from, 0)
		log.Printf("failed fetching %s (%s) from %s\n", target.Symbol, target.Exchange, ts.UTC().Format(time.RFC3339))
		log.Println(err.Error())
		return nil, throw.New(err, throw.ErrKindUnexpected)
	}

	return candles, nil
}

func fetchFuturesLatest(from int64, symbol string) ([]candlestick.Candle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	historyService := futuresClient.NewKlinesService()
	klines, err := historyService.Interval("1m").Symbol(symbol).Limit(99).StartTime(from * 1000).Do(ctx)
	cancel()
	if err != nil {
		return nil, err
	}
	candles := make([]candlestick.Candle, len(klines))
	for i, k := range klines {
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
	}
	return candles, nil
}

func fetchSpotLatest(from int64, symbol string) ([]candlestick.Candle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	historyService := spotClient.NewKlinesService()
	klines, err := historyService.Interval("1m").Symbol(symbol).Limit(99).StartTime(from * 1000).Do(ctx)
	cancel()
	if err != nil {
		return nil, err
	}
	candles := make([]candlestick.Candle, len(klines))
	for i, k := range klines {
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
	}
	return candles, nil
}
