package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jcawley/kyma-app-connector/internal"
	"github.com/jcawley/kyma-app-connector/pkg/connector"
	"github.com/jcawley/kyma-app-connector/pkg/mock"
)

func main() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", internal.IndexHandler)
	router.HandleFunc("/api/callTokenURL", connector.CallTokenURL)
	router.HandleFunc("/api/createSecureConnection", connector.CreateSecureConnection)
	router.HandleFunc("/api/getAppInfo", connector.GetAppInfo)
	router.HandleFunc("/api/sendAPISpec", connector.SendAPISpec)
	router.HandleFunc("/api/sendEventSpec", connector.SendEventSpec)
	router.HandleFunc("/mock/sendOrderCreatedEvent", mock.SendOrderCreatedEvent)

	log.Fatal(http.ListenAndServe(":8000", router))

	mock.StartMockServer()
}
