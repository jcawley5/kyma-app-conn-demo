package mock

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/jcawley/kyma-app-connector/pkg/connector"
	"github.com/jcawley/kyma-app-connector/pkg/utils"
)

//SendOrderCreatedEvent -
func SendOrderCreatedEvent(w http.ResponseWriter, r *http.Request) {

	client := connector.GetHTTPTLSClient()

	if client == nil {
		utils.ReturnError("No TLS Connection established", w)
		return
	}

	defer r.Body.Close()
	orderCode, err := ioutil.ReadAll(r.Body)

	if err != nil || len(orderCode) == 0 {
		log.Println("No orderCode provided... Sending default value...")
		orderCode = []byte("12345")
	}

	eventMessage := map[string]interface{}{
		"event-type":         "orderCreated",
		"event-type-version": "v1",
		"event-id":           "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"event-time":         time.Now().Format(time.RFC3339),
		"data": map[string]string{
			"orderCode": string(orderCode),
		},
	}

	eventBytes, _ := json.Marshal(eventMessage)

	eventURL := connector.GetEventURL()

	resp, err := client.Post(eventURL, "application/json", bytes.NewBuffer(eventBytes))

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		utils.ReturnError(err.Error(), w)
	} else {
		// evtTxt, _ := json.Marshal(eventMessage)
		AddOrderFromEvent(string(orderCode))
		utils.ReturnSuccess(string(respBody), w)
	}
}
