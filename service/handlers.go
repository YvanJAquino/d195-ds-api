package main

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
)

type Response struct {
	Message string    `json:"message,omitempty"`
	Time    time.Time `json:"time,omitempty"`
}

func WarmupRequestAccepted() *Response {
	return &Response{
		Message: "warmup request processed",
		Time:    time.Now(),
	}
}

func WarmupHandler(c *DataCache, l *GoogleCloudLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		json.NewEncoder(w).Encode(WarmupRequestAccepted())
		w.Header().Set("Content-Type", "application/json")
	}
}

func GetDataHandler(c *DataCache, l *GoogleCloudLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// if r.Method != http.MethodPost {
		// 	w.WriteHeader(http.StatusMethodNotAllowed)
		// 	return
		// }
		defer r.Body.Close()
		err := json.NewEncoder(w).Encode(c.Data())
		if err != nil {
			l.Log(EventCacheError, logging.Critical)
		}
	}
}

var (
	allowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodOptions}
	allowedHeaders = []string{"Origin", "Content-Type", "X-Requested-With"}
)

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		header := w.Header()
		header.Set("Access-Control-Allow-Origin", "*")
		for i := range allowedHeaders {
			header.Add("Access-Control-Allow-Headers", allowedHeaders[i])
		}
		for i := range allowedMethods {
			header.Add("Access-Control-Allow-Methods", allowedMethods[i])
		}
		next.ServeHTTP(w, r)
	}
}

func CORSMuxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(CORSMiddleware(next.ServeHTTP))
}
