package connector

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tidwall/sjson"
)

func init() {
	log.Print("rest connector initialize")
}

//CallTokenURL -
func (KymaConn *RestConnector) callTokenURL(oneTimeTokenURL string) ([]byte, error) {
	log.Println("CallTokenURL via rest...")

	resp, err := http.Get(oneTimeTokenURL)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBody, KymaConn)

	if err != nil {
		return nil, err
	}

	return respBody, nil
}

//SendCSRToKyma -
func (KymaConn *RestConnector) sendCSRToKyma(csr []byte) ([]byte, error) {
	log.Println("SendCSRToKyma via rest...")

	csrJSON := []byte(fmt.Sprintf("{\"csr\":\"%s\"}", base64.StdEncoding.EncodeToString(csr)))

	resp, err := http.Post(KymaConn.CsrURL, "application/json", bytes.NewBuffer(csrJSON))
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	csrRespData := &CSRConnectResponse{}
	err = json.Unmarshal([]byte(respBody), csrRespData)

	if err != nil {
		return nil, err
	}

	return []byte(csrRespData.ClientCrt), nil
}

//GetAppInfo - STEP 3
func (KymaConn *RestConnector) getAppInfo(TLSClient *http.Client) ([]byte, error) {
	log.Println("GetAppInfo via rest")

	resp, err := TLSClient.Get(KymaConn.API.InfoURL)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBody, KymaConn)

	return respBody, nil
}

//SendAPISpec - STEP 4
func (KymaConn *RestConnector) sendAPISpec(TLSClient *http.Client, APISpec []byte, hostURL []byte) ([]byte, error) {
	log.Println("SendAPISpec via rest")

	if KymaConn.URLs.MetadataURL == "" {
		return nil, errors.New("no MetadataURL exists")
	}

	json, err := sjson.Set(string(APISpec), "api.targetUrl", string(hostURL))

	if err != nil {
		return nil, err
	}

	resp, err := TLSClient.Post(KymaConn.URLs.MetadataURL, "application/json", bytes.NewBuffer([]byte(json)))

	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	log.Println(string(respBody))
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

//SendEventSpec -
func (KymaConn *RestConnector) sendEventSpec(TLSClient *http.Client, EventSpec []byte) ([]byte, error) {
	log.Println("SendEventSpec via rest")

	if KymaConn.URLs.MetadataURL == "" {
		return nil, errors.New("no MetadataURL exists")
	}

	resp, err := TLSClient.Post(KymaConn.URLs.MetadataURL, "application/json", bytes.NewBuffer(EventSpec))

	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (KymaConn *RestConnector) getCertificateSubject() string {
	log.Println("GetCertificateSubject")

	return KymaConn.Certificate.Subject
}

func (KymaConn *RestConnector) getEventURL() string {
	log.Println("getEventURL via rest")

	return KymaConn.API.EventsURL
}
