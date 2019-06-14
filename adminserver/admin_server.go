// admin_server handles administrative functions for bad server
package adminserver

import (
	"net/http"
	"strings"
)

type command string

const update command = "update"
const merge command = "merge"
const get command = "get"
const clear command = "clear"

type adminMessage struct {
	messageType command
}

type headerMessage struct {
	headers       http.Header
	returnChannel chan http.Header
	adminMessage
}

var headerCommands = make(chan headerMessage)
var defaultHeaders = make(map[string][]string)

func init() {
	go processHeaderCommands()
}

func processHeaderCommands() {
	for command := range headerCommands {
		switch command.messageType {
		case update:
			defaultHeaders = command.headers
			command.returnChannel <- defaultHeaders
		case get:
			command.returnChannel <- defaultHeaders
		default:
			command.returnChannel <- make(map[string][]string)
		}
	}
}

func RouteAdminCall(response http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/headers") {
		switch request.Method {
		case "GET":
			returnDefaultHeaders(response, request)
		case "POST":
			updateDefaultHeaders(response, request)
		case "DELETE":
			clearDefaultHeaders(response, request)
		default:
			response.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

var emptyHeaders = make(map[string][]string)

func GetCurrentHeaders() map[string][]string {
	returnChan := make(chan http.Header, 1)
	defer close(returnChan)
	headerCommands <- headerMessage{emptyHeaders, returnChan, adminMessage{get}}

	returnHeaders := <-returnChan

	return returnHeaders
}

func updateDefaultHeaders(response http.ResponseWriter, request *http.Request) {
	sendHeaderCommand(response, request, update)
}

func returnDefaultHeaders(response http.ResponseWriter, request *http.Request) {
	sendHeaderCommand(response, request, get)
}

func clearDefaultHeaders(response http.ResponseWriter, request *http.Request) {
	request.Header = make(map[string][]string)
	updateDefaultHeaders(response, request)
}

func sendHeaderCommand(response http.ResponseWriter, request *http.Request, messageType command) {
	returnChan := make(chan http.Header, 1)
	defer close(returnChan)

	headerCommands <- headerMessage{request.Header, returnChan, adminMessage{messageType}}

	headerResponse := <-returnChan

	for key, value := range headerResponse {
		response.Header()[key] = value
	}
}
