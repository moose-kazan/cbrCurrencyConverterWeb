package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"skfmod39/internal/currency"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	wg       sync.WaitGroup
	chanDone chan int
	rates    *currency.CurrencyList
)

func updateRates() {
	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(600 * time.Second)
	defer ticker.Stop()

	for {
		rates.Fetch()
		select {
		case <-chanDone:
			{
				return
			}
		case <-ticker.C:
			{
				fmt.Println("Time to fetch currencies!")
			}
		}
	}
}

func webApiCurrencyList(rw http.ResponseWriter, r *http.Request) {
	type RateItem struct {
		ISOCode string
		Name    string
	}
	var result []RateItem
	for _, v := range rates.Rates {
		var item = RateItem{
			ISOCode: v.GetISOCode(),
			Name:    v.GetName(),
		}
		result = append(result, item)
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(result)
}

func webApiCurrencyConvert(rw http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "application/json")
	rate, err := rates.Convert(params["from"], params["to"])
	_ = rate
	_ = err
	var result map[string]string = make(map[string]string)
	if err != nil {
		result["error"] = err.Error()
	} else {
		result["from"] = params["from"]
		result["to"] = params["to"]
		result["direct"] = fmt.Sprintf("%f", rate)
		result["reverse"] = fmt.Sprintf("%f", 1/rate)
	}
	json.NewEncoder(rw).Encode(result)
}

func webServerRun() {
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("./webroot/")))
	r.HandleFunc("/api/currency/list", webApiCurrencyList)
	r.HandleFunc("/api/currency/convert/{from:[A-Z]{3}}/{to:[A-Z]{3}}", webApiCurrencyConvert)
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func main() {
	chanDone = make(chan int)
	defer close(chanDone)

	rates = currency.New()

	go updateRates()
	webServerRun()

	wg.Wait()
}
