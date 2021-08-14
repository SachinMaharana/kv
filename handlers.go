package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
)

type Payload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeysResponse struct {
	Keys []string `json:"keys"`
}

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
}

func (app *application) getKey(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	key := params["key"]

	val, err := app.db.Get(key)
	if err == redis.Nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		app.logger.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s\n", val)
}

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	suffix := r.URL.Query().Get("suffix")

	if prefix == "" && suffix == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var searchKey string
	if prefix != "" {
		searchKey = fmt.Sprintf("%s*", prefix)
	} else {
		searchKey = fmt.Sprintf("*%s", suffix)
	}

	keys, err := app.db.Search(searchKey)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	var p KeysResponse
	p.Keys = keys

	// what should be the return response?
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (app *application) setKey(w http.ResponseWriter, r *http.Request) {
	var p Payload
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: Check for existence in db?
	err = app.db.Set(p.Key, p.Value)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// keys.WithLabelValues("redis").Set(float64(app.db.TotalKeys()))
	w.WriteHeader(http.StatusOK)
}
