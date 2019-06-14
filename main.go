package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"bad-server/adminserver"
	"bad-server/badness"
)

var port int
var adminPort int

type mainHandler struct{}

func (mainHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	mergeDefaultHeadersToRequest(request)
	pipeline := badness.GetResponsePipeline(request)
	for _, responseHandler := range pipeline {
		responseHandler(response)
	}
}

type adminHandler struct{}

func (adminHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	mergeDefaultHeadersToRequest(request)
	adminserver.RouteAdminCall(response, request)
	defer request.Body.Close()
}

func init() {
	flag.IntVar(&port, "port", 7865, "The port to listen on")
	flag.IntVar(&adminPort, "adminPort", 7866, "The port for admin functions")
}

func main() {
	flag.Parse()

	// use different server multiplexers for each server, to avoid path conflicts
	mainServerMux := http.NewServeMux()
	mainServerMux.Handle("/", &mainHandler{})
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mainServerMux))
	}()

	adminServerMux := http.NewServeMux()
	adminServerMux.Handle("/", &adminHandler{})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", adminPort), adminServerMux))
}

func mergeDefaultHeadersToRequest(request *http.Request) {
	defaultHeaders := adminserver.GetCurrentHeaders()

	for key, value := range defaultHeaders {
		if _, found := request.Header[key]; !found {
			request.Header[key] = value
		}
	}
}
