package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"skfmod39/internal/currency"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	wg         sync.WaitGroup
	chanDone   chan int
	chanErrors chan string
	rates      *currency.CurrencyList
)

func processErrors() {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case err := <-chanErrors:
			{
				os.Stderr.WriteString(fmt.Sprintf(
					"[%s] %s\n",
					time.Now().Format(time.RFC3339),
					err,
				))
			}
		case <-chanDone:
			{
				return
			}
		}
	}
}

func updateRates() {
	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(UPDATE_INTERVAL * time.Second)
	defer ticker.Stop()

	for {
		fmt.Println("Update rates")
		var err error
		err = rates.Fetch()
		if err != nil {
			chanErrors <- fmt.Sprintf("Can't update rates: %s", err.Error())
		}
		select {
		case <-chanDone:
			{
				return
			}
		case <-ticker.C:
			{

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
	for _, v := range rates.GetRates() {
		var item = RateItem{
			ISOCode: v.GetISOCode(),
			Name:    v.GetName(),
		}
		result = append(result, item)
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "application/json")
	if result == nil {
		json.NewEncoder(rw).Encode(map[string]string{"error": "No rates found!"})
	} else {
		json.NewEncoder(rw).Encode(result)
	}
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
		Addr:         LISTEN_SOCKET,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func main() {
	chanDone = make(chan int)
	chanErrors = make(chan string)
	defer close(chanDone)
	defer close(chanErrors)

	rates = currency.New()

	go processErrors()
	go updateRates()
	webServerRun()

	wg.Wait()
}
