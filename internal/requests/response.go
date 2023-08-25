package requests

import (
	"encoding/gob"
	"encoding/json"
	"net/http"
	"strings"
)

func sendAsJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func sendAsBinary(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_ = gob.NewEncoder(w).Encode(data)
}

func SendResponse(w http.ResponseWriter, r *http.Request, data interface{}) {

	// try to satisfy accept header
	accepts := r.Header.Get("Accept")

	// send as json when nothing is specified
	if accepts == "" {
		sendAsJSON(w, data)
		return
	}

	// send as json when json is requested
	if strings.Index(accepts, "application/json") != -1 {
		sendAsJSON(w, data)
		return
	}

	// send as gob when binary is requested
	if strings.Index(accepts, "application/octet-stream") != -1 {
		sendAsBinary(w, data)
		return
	}

	// send as json when any is requested
	if strings.Index(accepts, "*/*") != -1 {
		sendAsJSON(w, data)
		return
	}

	// deny other types
	w.WriteHeader(http.StatusNotAcceptable)

}
