package connector

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"

	cert "github.com/jcawley/kyma-app-connector/pkg/certificate"
	"github.com/jcawley/kyma-app-connector/pkg/utils"
)

//Connector -
type Connector interface {
	callTokenURL(string) ([]byte, error)
	sendCSRToKyma([]byte) ([]byte, error)
	getAppInfo(*http.Client) ([]byte, error)
	sendAPISpec(*http.Client, []byte, []byte) ([]byte, error)
	sendEventSpec(*http.Client, []byte) ([]byte, error)
	getCertificateSubject() string
	getEventURL() string
}

type apiConfig struct {
	AssetsDir           string
	kc                  Connector
	KymaClientCrtExists bool
	HTTPTLSClient       *http.Client
	ConnectionType      string
	KeyLength           int
}

var config *apiConfig

const appTypeRest string = "REST"
const appTypeGraphQL string = "GraphQL"

func init() {
	log.Println("init called")

	config = &apiConfig{}

	config.setAssetsDir()
	config.KymaClientCrtExists = false

	log.Println(config.AssetsDir)
}

//will generate a rest or graphql connection based on the tokenData
func initConnectionType(tokenData string) {

	_, err := base64.StdEncoding.DecodeString(tokenData)
	if err == nil {
		config.kc = &GraphQLConnector{}
		config.ConnectionType = appTypeGraphQL
		config.KeyLength = 4096
	} else {
		config.kc = &RestConnector{}
		config.ConnectionType = appTypeRest
		config.KeyLength = 2048
	}
	log.Printf("Initialized connection of type: %s", config.ConnectionType)
}

//CallTokenURL - STEP 1
func CallTokenURL(w http.ResponseWriter, r *http.Request) {
	log.Println("CallTokenURL")

	defer r.Body.Close()
	tokenData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ReturnError("Could not read the body text", w)
		return
	}

	tokenDataStr := string(tokenData)

	initConnectionType(tokenDataStr)

	resp, err := config.kc.callTokenURL(tokenDataStr)

	if err != nil {
		utils.ReturnError("Could not call the Token URL", w)
	} else {
		utils.ReturnSuccess(string(resp), w)
	}
}

//CreateSecureConnection - STEP 2
func CreateSecureConnection(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateSecureConnection")

	if config.kc == nil {
		utils.ReturnError("No connection has been established", w)
		return
	}

	subject := config.kc.getCertificateSubject()

	if subject == "" {
		utils.ReturnError("No Certificate Subject found", w)
		return
	}

	KymaCerts, err := cert.GenerateCSR(subject, config.KeyLength)

	if err != nil {
		utils.ReturnError("Could not generate the CSR", w)
		return
	}
	ioutil.WriteFile(config.AssetsDir+"/kymacerts/cert.csr", KymaCerts.CSR, 0644)

	clientCrt, err := config.kc.sendCSRToKyma(KymaCerts.CSR)

	if err != nil {
		utils.ReturnError(err.Error(), w)
	} else {
		config.saveTLSCerts(clientCrt, KymaCerts.PrivateKey)
		config.setTLSClient()
		utils.ReturnSuccess("Secure TLS Connection has been established", w)
	}
}

//GetAppInfo - STEP 3 REST
func GetAppInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetAppInfo")

	if config.HTTPTLSClient == nil {
		utils.ReturnError("No TLS Connection established", w)
		return
	}

	resp, err := config.kc.getAppInfo(config.HTTPTLSClient)

	if err != nil {
		utils.ReturnError(err.Error(), w)
	} else {
		if config.ConnectionType == appTypeRest {
			ioutil.WriteFile(config.AssetsDir+"/kyma-info-url-details.json", resp, 0644)
		}
		utils.ReturnSuccess(string(resp), w)
	}
}

//SendAPISpec - STEP 4
func SendAPISpec(w http.ResponseWriter, r *http.Request) {
	log.Println("SendAPISpec")

	if config.HTTPTLSClient == nil {
		utils.ReturnError("No TLS Connection established", w)
		return
	}

	var APISpec []byte
	var err error
	if config.ConnectionType == appTypeRest {
		APISpec, err = ioutil.ReadFile(config.AssetsDir + "/spec-docs/api-rest.json")
	} else {
		APISpec, err = ioutil.ReadFile(config.AssetsDir + "/spec-docs/api-graphql.yaml")
	}

	defer r.Body.Close()
	hostURL, err := ioutil.ReadAll(r.Body)

	if err != nil || len(hostURL) == 0 {
		log.Println("No hostURL provided... Setting to https://localhost:8443")
		hostURL = []byte("https://localhost:8443")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resp, err := config.kc.sendAPISpec(config.HTTPTLSClient, APISpec, hostURL)

	if err != nil {
		utils.ReturnError(err.Error(), w)
	} else {
		utils.ReturnSuccess(string(resp), w)
	}

}

//SendEventSpec - STEP 5
func SendEventSpec(w http.ResponseWriter, r *http.Request) {
	log.Println("SendEventSpec")

	if config.HTTPTLSClient == nil {
		utils.ReturnError("No TLS Connection established", w)
		return
	}

	var EventSpec []byte
	var err error
	if config.ConnectionType == appTypeRest {
		EventSpec, err = ioutil.ReadFile(config.AssetsDir + "/spec-docs/event-rest.json")
	} else {
		EventSpec, err = ioutil.ReadFile(config.AssetsDir + "/spec-docs/event-graphql.yaml")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resp, err := config.kc.sendEventSpec(config.HTTPTLSClient, EventSpec)

	if err != nil {
		utils.ReturnError(err.Error(), w)
	} else {
		utils.ReturnSuccess(string(resp), w)
	}
}

//GetAssetsDir -
func GetAssetsDir() string {
	return config.AssetsDir
}

//GetEventURL -
func GetEventURL() string {
	return config.kc.getEventURL()
}

//GetHTTPTLSClient -
func GetHTTPTLSClient() *http.Client {
	return config.HTTPTLSClient
}

//determines directory location
func (config *apiConfig) setAssetsDir() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	config.AssetsDir = filepath.Join(path.Dir(filename), "../../assets")
}

//saves the TLS certs for later use
func (config *apiConfig) saveTLSCerts(clientCrt []byte, privateKey []byte) {

	ioutil.WriteFile(config.AssetsDir+"/kymacerts/private.key", privateKey, 0644)

	decodedClientCrt, decodeErr := base64.StdEncoding.DecodeString(string(clientCrt))
	if decodeErr != nil {
		log.Fatalf("something went wrong decoding the ClientCrt")
	}
	clientDecodedBytes := []byte(string(decodedClientCrt))

	ioutil.WriteFile(config.AssetsDir+"/kymacerts/client.crt", clientDecodedBytes, 0644)

}

//Creates the TLS connection that all communication between the system will use
func (config *apiConfig) setTLSClient() {

	log.Println("Setting up TLS Client")
	srvCert, err := tls.LoadX509KeyPair(config.AssetsDir+"/kymacerts/client.crt", config.AssetsDir+"/kymacerts/private.key")
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile(config.AssetsDir + "/kymacerts/client.crt")
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	config.HTTPTLSClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{srvCert},
				InsecureSkipVerify: true,
			},
		},
	}
}
