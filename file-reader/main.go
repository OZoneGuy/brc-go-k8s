package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const CHUNK_SIZE int = 64 * 1024 * 1024

var PARSER_URL = os.Getenv("PARSER_URL")

func main() {
	res := make(chan map[string]cityTemperatureInfo)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /start", func(w http.ResponseWriter, r *http.Request) {
		// TODO: set file path
		filePath := "tmp"
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Could not open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		buf := make([]byte, CHUNK_SIZE)
		leftover := make([]byte, CHUNK_SIZE)
		for {
			readTotal, err := file.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				err = errors.Wrap(err, "Failed to read from file")
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			buf = buf[:readTotal]

			toSend := make([]byte, readTotal)
			copy(toSend, buf)

			lastNewLineIndex := bytes.LastIndex(buf, []byte{'\n'})

			toSend = append(leftover, buf[:lastNewLineIndex+1]...)
			leftover = make([]byte, len(buf[lastNewLineIndex+1:]))
			copy(leftover, buf[lastNewLineIndex+1:])

			// NOTE: Might need to make it a coroutine
			// Need to get the total result to the user somehow...
			resp, err := http.Post(PARSER_URL, "application/octet-stream", bytes.NewReader(toSend))
			if resp.StatusCode >= 300 {
				err = errors.Wrap(err, "Failed to make request to parser MS")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// NOTE: We could have a channel that waits on another MS to make a request
		// here and return the final value. Could be a solution to returning the
		// result.

		result := <-res
		resBytes, err := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(resBytes)
	})

	mux.HandleFunc("POST /complete", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request", http.StatusBadRequest)
			return
		}
		result := make(map[string]cityTemperatureInfo)
		err = json.Unmarshal(body, &result)
		if err != nil {
			http.Error(w, "Failed to marshal request", http.StatusBadRequest)
			return
		}
		res <- result

		w.WriteHeader(http.StatusAccepted)
		w.Write(nil)
	})

	server := &http.Server{
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
