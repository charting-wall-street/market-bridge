package web

import (
	"github.com/godoji/candlestick"
	"github.com/gorilla/mux"
	"marlin/internal/arbiter"
	"marlin/internal/requests"
	"marlin/internal/throw"
	"net/http"
	"strconv"
)

type CandlesPayload struct {
	Candles []candlestick.Candle `json:"candles"`
}

func HandleGetLatest(w http.ResponseWriter, r *http.Request) {

	// Parse source parameter
	target, ok := candlestick.ParseSymbol(mux.Vars(r)["uuid"])
	if !ok {
		throw.HttpError(w, throw.ErrInvalidSymbol)
		return
	}

	// Parse from parameter
	from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		throw.HttpError(w, throw.ErrInvalidFromParameter)
		return
	}

	// Fetch candles
	candles, ex := arbiter.FetchLatest(target, from)
	if ex != nil {
		throw.HttpError(w, ex)
		return
	} else {
		requests.SendResponse(w, r, CandlesPayload{candles})
	}
}

func HandleGetHistorical(w http.ResponseWriter, r *http.Request) {

	// Parse source parameter
	target, ok := candlestick.ParseSymbol(mux.Vars(r)["uuid"])
	if !ok {
		throw.HttpError(w, throw.ErrInvalidSymbol)
		return
	}

	// Parse from parameter
	from := int64(0)
	if s := r.URL.Query().Get("from"); s != "" {
		var err error
		from, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			throw.HttpError(w, throw.ErrInvalidFromParameter)
			return
		}
	}

	// Parse from parameter
	interval := int64(0)
	if s := r.URL.Query().Get("interval"); s != "" {
		var err error
		interval, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			throw.HttpError(w, throw.ErrInvalidInterval)
			return
		}
	} else {
		throw.HttpError(w, throw.ErrInvalidInterval)
		return
	}

	// Fetch candles
	candles, ex := arbiter.FetchHistorical(target, from, interval)
	if ex != nil {
		throw.HttpError(w, ex)
		return
	} else {
		requests.SendResponse(w, r, CandlesPayload{candles})
	}
}

func HandleGetInfo(w http.ResponseWriter, r *http.Request) {
	info := arbiter.ExchangeInfo()
	requests.SendResponse(w, r, info)
}
