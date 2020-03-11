package connector

//RestConnector -
type RestConnector struct {
	CsrURL      string `json:"csrUrl"`
	API         `json:"api"`
	Certificate `json:"certificate"`
	URLs        `json:"urls"`
}

//GraphQLConnector -
type GraphQLConnector struct {
	ConnectorURL          string                    `json:"connectorURL"`
	Token                 string                    `json:"token"`
	GraphQLAPIResp        GraphQLAPI                `json:"data"`
	Certificate           Certificate               `json:"certificate"`
	CSRConnectGraphQLResp CSRConnectGraphQLResponse `json:"data"`
	AppID                 AppID                     `json:"data"`
}

//API -
type API struct {
	EventsInfoURL   string `json:"eventsInfoUrl"`
	EventsURL       string `json:"eventsUrl"`
	MetadataURL     string `json:"metadataUrl"`
	InfoURL         string `json:"infoUrl"`
	CertificatesURL string `json:"certificatesUrl"`
}

//ClientIdentity -
type ClientIdentity struct {
	Application string `json:"application"`
}

//URLs -
type URLs struct {
	MetadataURL   string `json:"metadataUrl"`
	EventsURL     string `json:"eventsUrl"`
	RenewCertURL  string `json:"renewCertUrl"`
	RevokeCertURL string `json:"revokeCertUrl"`
}

//Certificate -
type Certificate struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

//CSRConnectResponse contains crt for tls communication
type CSRConnectResponse struct {
	Crt       string `json:"crt"`
	ClientCrt string `json:"clientCrt"`
	CaCrt     string `json:"caCrt"`
}

//CSRConnectGraphQLResponse -
type CSRConnectGraphQLResponse struct {
	Result struct {
		CertificateChain  string `json:"certificateChain"`
		CaCertificate     string `json:"caCertificate"`
		ClientCertificate string `json:"clientCertificate"`
	} `json:"result"`
}

//GraphQLAPI -
type GraphQLAPI struct {
	Result struct {
		Token struct {
			Token string `json:"token"`
		} `json:"token"`
		CertificateSigningRequestInfo `json:"certificateSigningRequestInfo"`
		ManagementPlaneInfo           struct {
			DirectorURL                    string `json:"directorURL"`
			CertificateSecuredConnectorURL string `json:"certificateSecuredConnectorURL"`
		} `json:"managementPlaneInfo"`
	} `json:"result"`
}

//CertificateSigningRequestInfo -
type CertificateSigningRequestInfo struct {
	Subject      string `json:"subject"`
	KeyAlgorithm string `json:"keyAlgorithm"`
}

//AppID -
type AppID struct {
	Viewer struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"viewer"`
}
