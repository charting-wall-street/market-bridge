package unicorn

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/godoji/candlestick"
	"io"
	"marlin/internal/config"
	"marlin/internal/throw"
	"net/http"
	"strconv"
	"time"
)

func FetchHistorical(target candlestick.AssetIdentifier) ([]candlestick.Candle, throw.Exception) {
	candles, err := fetchHistoricalRaw(target)
	if err != nil {
		return nil, throw.New(err, throw.ErrKindUnexpected)
	}
	return candles, nil
}

func fetchHistoricalRaw(target candlestick.AssetIdentifier) ([]candlestick.Candle, error) {
	url := fmt.Sprintf("%s/eod/%s.%s?api_token=%s&period=d", config.UnicornAPI, target.Symbol, target.Exchange, config.ServiceConfig().UnicornKey())
	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if req.StatusCode != http.StatusOK {
		return nil, errors.New(req.Status)
	}

	defer req.Body.Close()

	reader := csv.NewReader(req.Body)
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	currentDateSet := false
	currentDate := time.Now()
	candles := make([]candlestick.Candle, 0)
	for true {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		date, err := time.ParseInLocation("2006-01-02", line[0], time.UTC)
		if err != nil {
			return nil, err
		}
		priceOpen, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			return nil, err
		}
		priceHigh, err := strconv.ParseFloat(line[2], 64)
		if err != nil {
			return nil, err
		}
		priceLow, err := strconv.ParseFloat(line[3], 64)
		if err != nil {
			return nil, err
		}
		priceClose, err := strconv.ParseFloat(line[4], 64)
		if err != nil {
			return nil, err
		}
		volume, err := strconv.ParseFloat(line[6], 64)
		if err != nil {
			return nil, err
		}

		if !currentDateSet {
			currentDateSet = true
			currentDate = date
		} else {
			currentDate = currentDate.AddDate(0, 0, 1)
			for date != currentDate {
				candles = append(candles, candlestick.Candle{
					Time:    currentDate.Unix(),
					Missing: true,
				})
				currentDate = currentDate.AddDate(0, 0, 1)
			}
		}

		candles = append(candles, candlestick.Candle{
			Open:           priceOpen,
			High:           priceHigh,
			Low:            priceLow,
			Close:          priceClose,
			Volume:         volume,
			TakerVolume:    0,
			NumberOfTrades: 0,
			Time:           date.Unix(),
			Missing:        false,
		})
	}

	return candles, nil
}
