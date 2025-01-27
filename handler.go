package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func healthEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	cfg.metricsHandler(w, r)
}

func (cfg *apiConfig) validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}
	type errJson struct {
		Error string `json:"error"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := errJson{
			Error: fmt.Sprintf("Decoding error %v", err),
		}

		dat, e := json.Marshal(respBody)
		if err != nil {
			log.Printf("Marshalling error [%v] for decoding error: %v", e, err)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(dat)
		return
	}

	if len(params.Body) > 140 {
		respBody := errJson{
			Error: "Error: Message exceeds character limit",
		}

		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Marshalling error [%v] for message exceeding char limit", err)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	respBody := returnVals{
		Valid: true,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		respBody := errJson{
			Error: fmt.Sprintf("Marshalling error: %v", err),
		}

		dat, e := json.Marshal(respBody)
		if e != nil {
			log.Printf("Marshalling error [%v] for marshallin error : %v", e, err)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(dat)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}
