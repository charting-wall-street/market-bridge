package binance

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/godoji/candlestick"
	"log"
	"os"
	"strconv"
	"time"
)

var futuresClient = futures.NewClient("", "")
var spotClient = binance.NewClient("", "")

const fetchLimit = 1000

func klineToCandle(o string, h string, l string, c string, v string, tn int64, tv string, t int64) candlestick.Candle {
	candleOpen, _ := strconv.ParseFloat(o, 64)
	candleHigh, _ := strconv.ParseFloat(h, 64)
	candleLow, _ := strconv.ParseFloat(l, 64)
	candleClose, _ := strconv.ParseFloat(c, 64)
	candleVolume, _ := strconv.ParseFloat(v, 64)
	candleNumberTrades := tn
	candleTakerVolume, _ := strconv.ParseFloat(tv, 64)
	candleTime := t
	return candlestick.Candle{
		Open:           candleOpen,
		High:           candleHigh,
		Low:            candleLow,
		Close:          candleClose,
		Volume:         candleVolume,
		NumberOfTrades: candleNumberTrades,
		TakerVolume:    candleTakerVolume,
		Time:           candleTime,
	}
}

func futureToSpot(candles []*futures.Kline) []*binance.Kline {
	results := make([]*binance.Kline, len(candles))
	for i, candle := range candles {
		results[i] = &binance.Kline{
			OpenTime:                 candle.OpenTime,
			Open:                     candle.Open,
			High:                     candle.High,
			Low:                      candle.Low,
			Close:                    candle.Close,
			Volume:                   candle.Volume,
			CloseTime:                candle.CloseTime,
			QuoteAssetVolume:         candle.QuoteAssetVolume,
			TradeNum:                 candle.TradeNum,
			TakerBuyBaseAssetVolume:  candle.TakerBuyBaseAssetVolume,
			TakerBuyQuoteAssetVolume: candle.QuoteAssetVolume,
		}
	}
	return results
}

func fillMissingCandles(candles []*binance.Kline, from time.Time, symbol string) []binance.Kline {

	results := make([]binance.Kline, 1000)
	i := 0
	filled := 0

	lastOpen := from.UnixMilli() - time.Minute.Milliseconds()
	for _, candle := range candles {

		// downtime candles
		for candle.OpenTime-lastOpen != time.Minute.Milliseconds() {

			lastOpen += time.Minute.Milliseconds()

			results[i] = binance.Kline{
				OpenTime:  lastOpen,
				CloseTime: lastOpen + 60000 - 1, // 1 minute in milliseconds
			}
			filled += 1
			i++

			if i == 1000 {

				requestTimeStamp := from.Unix()
				responseTimeStamp := candles[0].OpenTime / 1000

				log.Printf("%s dropped candles @%d: %d filled \n", symbol, from.Unix()/60/5000, filled)

				if responseTimeStamp < requestTimeStamp {
					log.Println("response is invalid, candles start earlier than requested")
					log.Printf("expected %d got %d instead\n", requestTimeStamp, responseTimeStamp)
					os.Exit(1)
				} else if responseTimeStamp > requestTimeStamp {
					// first available candle is returned when requesting a timestamp before first available
				} else {
					// occurs when some candles are not available but the first candle is
				}

				return results
			}
		}

		lastOpen = candle.OpenTime
		results[i] = *candle
		i++

		if i == 1000 {
			break
		}
	}

	// fill time in remaining candles if not enough candles were returned
	for i != 1000 {
		lastOpen += time.Minute.Milliseconds()
		results[i] = binance.Kline{
			OpenTime:  lastOpen,
			CloseTime: lastOpen + 60000 - 1,
		}
		i++
	}

	return results
}
