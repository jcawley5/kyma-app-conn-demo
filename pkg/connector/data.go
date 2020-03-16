package connector

//REST API

type restConnector struct {
	CsrURL      string      `json:"csrUrl"`
	API         api         `json:"api"`
	Certificate certificate `json:"certificate"`
	Urls        urls        `json:"urls"`
}

type api struct {
	EventsInfoURL   string `json:"eventsInfoUrl"`
	EventsURL       string `json:"eventsUrl"`
	MetadataURL     string `json:"metadataUrl"`
	InfoURL         string `json:"infoUrl"`
	CertificatesURL string `json:"certificatesUrl"`
}

type clientIdentity struct {
	Application string `json:"application"`
}

type urls struct {
	MetadataURL   string `json:"metadataUrl"`
	EventsURL     string `json:"eventsUrl"`
	RenewCertURL  string `json:"renewCertUrl"`
	RevokeCertURL string `json:"revokeCertUrl"`
}

//Certificate -
type certificate struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

//CSRConnectResponse contains crt for tls communication
type csrConnectResponse struct {
	Crt       string `json:"crt"`
	ClientCrt string `json:"clientCrt"`
	CaCrt     string `json:"caCrt"`
}

//GRAPHQL API

//GraphQLConnector -
type graphQLConnector struct {
	ConnectorURL          string                    `json:"connectorURL"`
	Token                 string                    `json:"token"`
	GraphQLAPIResp        graphQLAPI                `json:"data"`
	Certificate           certificate               `json:"certificate"`
	CsrConnectGraphQLResp csrConnectGraphQLResponse `json:"data"`
	AppID                 appID                     `json:"data"`
	EventsURL             eventsURL                 `json:"data"`
}

//GraphQLAPI -
type graphQLAPI struct {
	Result struct {
		Token struct {
			Token string `json:"token"`
		} `json:"token"`
		CertificateSigningRequestInfo certificateSigningRequestInfo `json:"certificateSigningRequestInfo"`
		ManagementPlaneInfo           struct {
			DirectorURL                    string `json:"directorURL"`
			CertificateSecuredConnectorURL string `json:"certificateSecuredConnectorURL"`
		} `json:"managementPlaneInfo"`
	} `json:"result"`
}

//CertificateSigningRequestInfo -
type certificateSigningRequestInfo struct {
	Subject      string `json:"subject"`
	KeyAlgorithm string `json:"keyAlgorithm"`
}

//CSRConnectGraphQLResponse -
type csrConnectGraphQLResponse struct {
	Result struct {
		CertificateChain  string `json:"certificateChain"`
		CaCertificate     string `json:"caCertificate"`
		ClientCertificate string `json:"clientCertificate"`
	} `json:"result"`
}

//AppID -
type appID struct {
	Viewer struct {
		ID         string `json:"id"`
		ViewerType string `json:"type"`
	} `json:"viewer"`
}

//EventsURL -
type eventsURL struct {
	Application struct {
		EventingConfiguration struct {
			DefaultURL string `json:"defaultURL"`
		} `json:"eventingConfiguration"`
	} `json:"application"`
}

type specDefinitionResp struct {
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}
