package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

var COLLECTOR_URL = os.Getenv("COLLECTOR_URL")

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("POST /parse", func(w http.ResponseWriter, r *http.Request) {

		result := make(map[string]cityTemperatureInfo)
		var start int
		var city string

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Respond to caller
		w.WriteHeader(http.StatusAccepted)
		_, err = w.Write([]byte{})
		if err != nil {
			// Failed to respond to caller
			log.Fatalf("Failed to respond to caller: %v", err)
			panic(err)
		}

		stringBuf := string(body)
		for index, char := range stringBuf {
			switch char {
			case ';':
				city = stringBuf[start:index]
				start = index + 1
			case '\n':
				if (index-start) > 1 && len(city) != 0 {
					temp := customStringToIntParser(stringBuf[start:index])
					start = index + 1

					if val, ok := result[city]; ok {
						val.Count++
						val.Sum += temp
						if temp < val.Min {
							val.Min = temp
						}

						if temp > val.Max {
							val.Max = temp
						}
						result[city] = val
					} else {
						result[city] = cityTemperatureInfo{
							Count: 1,
							Min:   temp,
							Max:   temp,
							Sum:   temp,
						}
					}
					city = ""
				}
			}
		}

		// Send to last MS
		resBytes, err := json.Marshal(result)
		http.Post(COLLECTOR_URL, "application/json", bytes.NewReader(resBytes))
	})

	server := &http.Server{
		Addr:    ":80",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

func customStringToIntParser(input string) (output int64) {
	var isNegativeNumber bool
	if input[0] == '-' {
		isNegativeNumber = true
		input = input[1:]
	}

	switch len(input) {
	case 3:
		output = int64(input[0])*10 + int64(input[2]) - int64('0')*11
	case 4:
		output = int64(input[0])*100 + int64(input[1])*10 + int64(input[3]) - (int64('0') * 111)
	}

	if isNegativeNumber {
		return -output
	}
	return
}

type cityTemperatureInfo struct {
	Count int64 `json:"count"`
	Min   int64 `json:"min"`
	Max   int64 `json:"max"`
	Sum   int64 `json:"sum"`
}
