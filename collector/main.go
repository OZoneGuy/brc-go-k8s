package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

var PARENT_URL = os.Getenv("READER_URL")

var mapOfTemp sync.Map = sync.Map{}

// 1 Billion
const MAX_ENTRIES = 1e9

var currentEntries = atomic.Uint64{}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /collect", func(w http.ResponseWriter, r *http.Request) {
		// process all city temperatures derived after processing the file chunks
		fmt.Println("Recieved request")
		var resultStream struct {
			Count  int                            `json:"count"`
			Cities map[string]cityTemperatureInfo `json:"cities"`
		}
		err := json.NewDecoder(r.Body).Decode(&resultStream)
		if err != nil {
			fmt.Printf("Failed to decode body: %v", err)
			err = errors.Wrap(err, "Failed to decode body")
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		fmt.Println("Starting processing loop")
		for city, tempInfo := range resultStream.Cities {
			if val, ok := mapOfTemp.Load(city); ok {
				cityInfo := val.(cityTemperatureInfo)
				cityInfo.Count += tempInfo.Count
				cityInfo.Sum += tempInfo.Sum
				if tempInfo.Min < cityInfo.Min {
					cityInfo.Min = tempInfo.Min
				}

				if tempInfo.Max > cityInfo.Max {
					cityInfo.Max = tempInfo.Max
				}
				mapOfTemp.Store(city, cityInfo)
			} else {
				mapOfTemp.Store(city, tempInfo)
			}
		}

		// respond
		fmt.Println("Responding")
		fmt.Fprint(w, "Added entries")

		// add value and respond to main MS
		fmt.Println("Adding to total count")
		count := uint64(resultStream.Count)
		currentEntries.Add(count)

		fmt.Printf("Checking if we reached count: %d\n", currentEntries.Load())
		if currentEntries.Load() == MAX_ENTRIES {
			// Finished with all entries. Call to original MS and return result
			fmt.Printf("Reached max count: %d\n", currentEntries.Load())
			resMap := map[string]cityTemperatureInfo{}
			mapOfTemp.Range(func(key, value any) bool {
				resMap[key.(string)] = value.(cityTemperatureInfo)
				return true
			})
			resBody, err := json.Marshal(resMap)
			if err != nil {
				panic(err)
			}
			resp, err := http.Post(PARENT_URL, "application/json", bytes.NewReader(resBody))
			if err != nil || resp.StatusCode != 202 {
				fmt.Printf("Failed to call parent: %v - %v\n", err, resp)
			}
		}
	})

	server := http.Server{
		Addr:    ":80",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

type cityTemperatureInfo struct {
	Count int64 `json:"count"`
	Min   int64 `json:"min"`
	Max   int64 `json:"max"`
	Sum   int64 `json:"sum"`
}
