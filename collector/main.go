package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

const PARENT_URL = "file-loader-svc.default.svc.cluster.local"

var mapOfTemp sync.Map = sync.Map{}

// 1 Billion
const MAX_ENTRIES = 1e9

var currentEntries = atomic.Uint32{}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /collect", func(w http.ResponseWriter, r *http.Request) {
		// process all city temperatures derived after processing the file chunks
		var resultStream map[string]cityTemperatureInfo = make(map[string]cityTemperatureInfo)
		for city, tempInfo := range resultStream {
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
		fmt.Fprint(w, "Added entries")

		// add value and respond to main MS
		count := uint32(len(resultStream))
		currentEntries.Add(count)

		if currentEntries.Load() == MAX_ENTRIES {
			// Finished with all entries. Call to original MS and return result
			resMap := map[string]cityTemperatureInfo{}
			mapOfTemp.Range(func(key, value any) bool {
				resMap[key.(string)] = value.(cityTemperatureInfo)
				return true
			})
			resBody, err := json.Marshal(resMap)
			if err != nil {
				panic(err)
			}
			http.Post(PARENT_URL, "application/json", bytes.NewReader(resBody))
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
