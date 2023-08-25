package web

import (
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"log"
	"marlin/internal/config"
	"net/http"
)

func router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/market/{uuid}/historical", HandleGetHistorical).Methods("GET")
	r.HandleFunc("/market/{uuid}/latest", HandleGetLatest).Methods("GET")
	r.HandleFunc("/market/info", HandleGetInfo).Methods("GET")
	return r
}

func Start() {

	// Middleware and routes
	app := negroni.New(negroni.NewRecovery())
	app.UseHandler(router())

	// Setup server
	server := &http.Server{
		Addr:    ":" + config.ServiceConfig().Port(),
		Handler: app,
	}

	log.Printf("listening on port %s\n", config.ServiceConfig().Port())

	log.Fatal(server.ListenAndServe())

}
