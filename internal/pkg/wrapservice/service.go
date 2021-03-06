package wrapservice

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/airenas/tts-line/internal/pkg/wrapservice/api"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/gorilla/mux"

	"github.com/pkg/errors"
)

type (
	//WaveSynthesizer main sythesis processor
	WaveSynthesizer interface {
		Work(string) (string, error)
	}
	//Data is service operation data
	Data struct {
		Port      int
		Processor WaveSynthesizer
	}
)

//StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP TTS Line service at %d", data.Port)
	r := NewRouter(data)
	http.Handle("/", r)
	portStr := strconv.Itoa(data.Port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

//NewRouter creates the router for HTTP service
func NewRouter(data *Data) *mux.Router {
	router := mux.NewRouter()
	router.Methods("POST").Path("/synthesize").Handler(&synthesisHandler{data: data})
	return router
}

type synthesisHandler struct {
	data *Data
}

func (h *synthesisHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	goapp.Log.Infof("Request from %s", r.RemoteAddr)

	decoder := json.NewDecoder(r.Body)
	var inText api.Input
	err := decoder.Decode(&inText)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		goapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	resp, err := h.data.Processor.Work(inText.Text)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		goapp.Log.Error("Can't process. ", err)
		return
	}
	res := &api.Result{Data: resp}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&res)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		goapp.Log.Error(err)
	}
}
