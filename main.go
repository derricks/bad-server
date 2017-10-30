package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"bad-server/badness"
)

var port int

func init() {
	flag.IntVar(&port, "port", 7865, "The port to listen on")
}

func main() {
	flag.Parse()

	http.HandleFunc("/", handleBadServerRequest)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// handleBadServerRequest will delegate to appropriate other handlers to generate
// the appropriate bad response
func handleBadServerRequest(response http.ResponseWriter, request *http.Request) {
	pipeline := badness.GetResponsePipeline(request)
	for _, responseHandler := range pipeline {
		responseHandler(response)
	}
}
