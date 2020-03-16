package connector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/machinebox/graphql"
)

func init() {
	log.Print("graphql connector initialize")
}

//CallTokenURL - STEP 1
func (KymaConn *graphQLConnector) callTokenURL(tokenData string) ([]byte, error) {
	log.Println("GraphQLConnector callTokenURL")

	tokenDataDecoded, err := base64.StdEncoding.DecodeString(tokenData)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(tokenDataDecoded, &KymaConn)

	client := graphql.NewClient(KymaConn.ConnectorURL)

	req := graphql.NewRequest(`
		query {
			result: configuration {
				token {
					token
				}
				certificateSigningRequestInfo {
					subject
					keyAlgorithm
				}
				managementPlaneInfo {
					directorURL
					certificateSecuredConnectorURL
				}
			}
		}
	`)
	req.Header.Add("connector-token", KymaConn.Token)

	// define a Context for the request
	ctx := context.Background()

	// run it and capture the response
	if err := client.Run(ctx, req, &KymaConn.GraphQLAPIResp); err != nil {
		return nil, err
	}

	// fmt.Printf("%+v\n", KymaConn.GraphQLAPI)

	return []byte(fmt.Sprintf("%+v\n", KymaConn.GraphQLAPIResp)), nil
}

//SendCSRToKyma - STEP 2
func (KymaConn *graphQLConnector) sendCSRToKyma(csr []byte) ([]byte, error) {
	log.Println("SendCSRToKyma via via graphql...")

	client := graphql.NewClient(KymaConn.ConnectorURL)

	req := graphql.NewRequest(`
	mutation ($csrBase64: String!) {
		result: signCertificateSigningRequest(csr: $csrBase64) {
			certificateChain
			caCertificate
			clientCertificate
		}
	}
	`)

	csrBase64 := base64.StdEncoding.EncodeToString(csr)
	req.Var("csrBase64", csrBase64)
	req.Header.Add("connector-token", KymaConn.GraphQLAPIResp.Result.Token.Token)

	// define a Context for the request
	ctx := context.Background()

	// run it and capture the response
	// csrRespData := &CSRConnectGraphQLResponse{}
	if err := client.Run(ctx, req, &KymaConn.CsrConnectGraphQLResp); err != nil {
		return nil, err
	}

	// log.Println("GraphQLAPI Result..............")
	fmt.Printf("%+v\n", KymaConn)

	return []byte(KymaConn.CsrConnectGraphQLResp.Result.CertificateChain), nil
}

//GetAppInfo -
func (KymaConn *graphQLConnector) getAppInfo(TLSClient *http.Client) ([]byte, error) {

	client := graphql.NewClient(KymaConn.GraphQLAPIResp.Result.ManagementPlaneInfo.DirectorURL, graphql.WithHTTPClient(TLSClient))

	err := KymaConn.getAppID(client)
	if err != nil {
		return nil, err
	}

	err = KymaConn.getEventsURL(client)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("{AppID: %s, EventsURL: %s}", KymaConn.AppID.Viewer.ID, KymaConn.EventsURL.Application.EventingConfiguration.DefaultURL)), nil

}

func (KymaConn *graphQLConnector) getAppID(client *graphql.Client) error {

	req := graphql.NewRequest(`
		query {
			viewer {
			id
			type
			}
		}
		`)

	ctx := context.Background()

	if err := client.Run(ctx, req, &KymaConn.AppID); err != nil {
		return err
	}
	fmt.Printf("%+v\n", KymaConn)

	return nil

}

func (KymaConn *graphQLConnector) getEventsURL(client *graphql.Client) error {
	req := graphql.NewRequest(`
	query($appId: ID!) {
		application (id : $appId){
		  eventingConfiguration{
			defaultURL
		  }
		}
	  }
	`)

	req.Var("appId", KymaConn.AppID.Viewer.ID)

	ctx := context.Background()

	if err := client.Run(ctx, req, &KymaConn.EventsURL); err != nil {
		return err
	}
	fmt.Printf("%+v\n", KymaConn)

	return nil
}

//SendEventSpec -
func (KymaConn *graphQLConnector) sendEventSpec(TLSClient *http.Client, eventSpec []byte) ([]byte, error) {
	log.Println("SendEventMetadata via graphql...")

	if KymaConn.AppID.Viewer.ID == "" {
		return nil, errors.New("no AppId exists")
	}

	client := graphql.NewClient(KymaConn.GraphQLAPIResp.Result.ManagementPlaneInfo.DirectorURL, graphql.WithHTTPClient(TLSClient))

	req := graphql.NewRequest(`
	mutation ($appID: ID!, $eventSpec: CLOB!){
		result: addEventDefinition(
		  	applicationID: $appID
		  	in: {
				name: "Sample Order Event - MP"
				spec: {
			  		type: ASYNC_API
			  		format: YAML
			  		data: $eventSpec
			  	}
		  	} 	
		){
			id
		  }
		}
	`)

	req.Var("appID", KymaConn.AppID.Viewer.ID)
	req.Var("eventSpec", string(eventSpec))

	var specDefResp specDefinitionResp

	ctx := context.Background()

	if err := client.Run(ctx, req, &specDefResp); err != nil {
		fmt.Printf(err.Error())
		return nil, err
	}
	fmt.Printf("%+v\n", KymaConn)

	return []byte(fmt.Sprintf("{ID: %s}", specDefResp.Result.ID)), nil
}

//SendAPISpec -
func (KymaConn *graphQLConnector) sendAPISpec(TLSClient *http.Client, apiSpec []byte, hostURL []byte) ([]byte, error) {

	log.Println("SendAPIMetadata via graphql...")

	if KymaConn.AppID.Viewer.ID == "" {
		return nil, errors.New("no AppId exists")
	}

	client := graphql.NewClient(KymaConn.GraphQLAPIResp.Result.ManagementPlaneInfo.DirectorURL, graphql.WithHTTPClient(TLSClient))

	req := graphql.NewRequest(`
	mutation ($appID: ID!, $apiSpec: CLOB!, $hostURL: String!){
		result: addAPIDefinition(
		  	applicationID: $appID
		  	in: {
				name: "Sample Order API - MP"
				targetURL: $hostURL
				spec: {
			  		type: OPEN_API
			  		format: YAML
			  		data: $apiSpec
				},
				defaultAuth: {
					credential: {
						basic: {
							username: "user"
							password: "password"
						}
					},				
					additionalHeaders: {
						header1: ["header1value"]
					},
					additionalQueryParams: {
						query1: ["query1value"]
					} 	
				}
			}
		){
			id
		  }
		}
	`)

	req.Var("appID", KymaConn.AppID.Viewer.ID)
	req.Var("apiSpec", string(apiSpec))
	req.Var("hostURL", string(hostURL))

	var specDefResp specDefinitionResp

	ctx := context.Background()

	if err := client.Run(ctx, req, &specDefResp); err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", KymaConn)

	return []byte(fmt.Sprintf("{ID: %s}", specDefResp.Result.ID)), nil
}

func (KymaConn *graphQLConnector) getCertificateSubject() string {
	log.Println("GetCertificateSubject")

	return KymaConn.GraphQLAPIResp.Result.CertificateSigningRequestInfo.Subject
}

func (KymaConn *graphQLConnector) getEventURL() string {
	log.Println("getEventURL via graphql")

	return KymaConn.EventsURL.Application.EventingConfiguration.DefaultURL
}
