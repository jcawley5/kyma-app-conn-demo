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
	AssetsDir        string
	kc               Connector
	HTTPTLSClient    *http.Client
	ConnectionType   string
	ConnectionStatus string
	KeyLength        int
}

var config *apiConfig

const appTypeRest string = "REST"
const appTypeGraphQL string = "GraphQL"
const notConnected string = "Not Connected"
const isConnected string = "Connected"

//GetConnectionStatus -
func GetConnectionStatus() string {

	return config.ConnectionStatus

}

func init() {
	log.Println("init api called")

	config = &apiConfig{}
	config.ConnectionStatus = notConnected
	config.setAssetsDir()
	// _ = config.setTLSClient()
}

//will generate a rest or graphql connection based on the tokenData
func initConnectionType(appType string) {

	if appType == appTypeGraphQL {
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

func getConnTypeByTokenData(tokenData string) string {
	var appType string
	_, err := base64.StdEncoding.DecodeString(tokenData)
	if err == nil {
		appType = appTypeGraphQL
	} else {
		appType = appTypeRest
	}
	return appType
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

	connType := getConnTypeByTokenData(tokenDataStr)
	initConnectionType(connType)

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

	crtChain, err := config.kc.sendCSRToKyma(KymaCerts.CSR)

	if err != nil {
		utils.ReturnError(err.Error(), w)
	} else {
		config.saveTLSCerts(crtChain, KymaCerts.PrivateKey)
		err := config.setTLSClient()
		if err != nil {
			utils.ReturnError("Could not establish a Secure TLS Connection", w)
		} else {
			utils.ReturnSuccess("Secure TLS Connection has been established", w)
		}

	}
}

//GetAppInfo - STEP 3
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
		log.Println("No hostURL provided... Setting to http://localhost:8000")
		hostURL = []byte("http://localhost:8000")
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
func (config *apiConfig) saveTLSCerts(crtChain []byte, privateKey []byte) {

	ioutil.WriteFile(config.AssetsDir+"/kymacerts/private.key", privateKey, 0644)

	decodedCrtChain, decodeErr := base64.StdEncoding.DecodeString(string(crtChain))
	if decodeErr != nil {
		log.Fatalf("something went wrong decoding the crtChain")
	}
	crtDecodedBytes := []byte(string(decodedCrtChain))

	ioutil.WriteFile(config.AssetsDir+"/kymacerts/crtChain.crt", crtDecodedBytes, 0644)

}

//Creates the TLS connection that all communication between the system will use
func (config *apiConfig) setTLSClient() error {

	log.Println("setTLSClient....")
	log.Println(config.AssetsDir)
	config.ConnectionStatus = notConnected

	keyPair, err := tls.LoadX509KeyPair(config.AssetsDir+"/kymacerts/crtChain.crt", config.AssetsDir+"/kymacerts/private.key")
	if err != nil {
		log.Println("setTLSClient: no keypair exists")
		return err
	}

	crtChain, err := ioutil.ReadFile(config.AssetsDir + "/kymacerts/crtChain.crt")
	if err != nil {
		log.Println("setTLSClient: no clientCrt exists")
		return err
	}

	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}

	caCertPool.AppendCertsFromPEM(crtChain)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{keyPair},
		RootCAs:            caCertPool,
	}
	tlsConfig.BuildNameToCertificate()

	config.HTTPTLSClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	log.Println("TLSClient has been set...")
	config.ConnectionStatus = isConnected
	return nil
}
