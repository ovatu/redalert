package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/ovatu/redalert/core"
	"github.com/rs/cors"
)

func Run(service *core.Service, port int, disableBrand bool, readOnly bool) {

	context := &appCtx{
		service: service,
		config: Config{
			disableBrand: disableBrand,
			readOnly: readOnly,
		},
	}

	box := rice.MustFindBox("assets")
	fs := http.FileServer(box.HTTPBox())

	router := mux.NewRouter()
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fs))
	router.Handle("/", fs)
	router.Handle("/api/put", appHandler{context, metricsReceiverHandler})

	router.Handle("/v1/stats", appHandler{context, statsHandler})
	router.Handle("/v1/checks/{check_id}/disable", appHandler{context, checkDisableHandler}).Methods("POST")
	router.Handle("/v1/checks/{check_id}/enable", appHandler{context, checkEnableHandler}).Methods("POST")
	router.Handle("/v1/checks/{check_id}/trigger", appHandler{context, checkTriggerHandler}).Methods("POST")

	router.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := cors.Default().Handler(router)
	err := http.ListenAndServe(":"+strconv.Itoa(port), handler)
	if err != nil {
		log.Fatal(err)
	}
}

type appCtx struct {
	service *core.Service
	config  Config
}

type Config struct {
	disableBrand bool
	readOnly bool
}

type appHandler struct {
	*appCtx
	h func(*appCtx, http.ResponseWriter, *http.Request)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.h(ah.appCtx, w, r)
}

func Respond(w http.ResponseWriter, data interface{}, code int) {
	response, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
