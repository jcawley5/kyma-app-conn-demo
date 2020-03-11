package connector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/machinebox/graphql"
)

func init() {
	log.Print("graphql connector initialize")
}

//CallTokenURL - STEP 1
func (KymaConn *GraphQLConnector) callTokenURL(tokenData string) ([]byte, error) {
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
func (KymaConn *GraphQLConnector) sendCSRToKyma(csr []byte) ([]byte, error) {
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
	if err := client.Run(ctx, req, &KymaConn.CSRConnectGraphQLResp); err != nil {
		return nil, err
	}

	// log.Println("GraphQLAPI Result..............")
	fmt.Printf("%+v\n", KymaConn)

	return []byte(KymaConn.CSRConnectGraphQLResp.Result.ClientCertificate), nil
}

//GetAppInfo -
func (KymaConn *GraphQLConnector) getAppInfo(TLSClient *http.Client) ([]byte, error) {

	client := graphql.NewClient(KymaConn.GraphQLAPIResp.Result.ManagementPlaneInfo.DirectorURL, graphql.WithHTTPClient(TLSClient))

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
		return nil, err
	}
	fmt.Printf("%+v\n", KymaConn)

	return []byte(fmt.Sprintf("{AppID: %s}", KymaConn.AppID.Viewer.ID)), nil

}

type specDefinitionResp struct {
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}

//SendEventSpec -
func (KymaConn *GraphQLConnector) sendEventSpec(TLSClient *http.Client, eventSpec []byte) ([]byte, error) {
	log.Println("SendEventMetadata via graphql...")

	client := graphql.NewClient(KymaConn.GraphQLAPIResp.Result.ManagementPlaneInfo.DirectorURL, graphql.WithHTTPClient(TLSClient))

	req := graphql.NewRequest(`
	mutation ($appID: ID!, $eventSpec: CLOB!){
		result: addEventDefinition(
		  	applicationID: $appID
		  	in: {
				name: "sample-app-event"
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
func (KymaConn *GraphQLConnector) sendAPISpec(TLSClient *http.Client, apiSpec []byte, hostURL []byte) ([]byte, error) {

	log.Println("SendAPIMetadata via graphql...")

	client := graphql.NewClient(KymaConn.GraphQLAPIResp.Result.ManagementPlaneInfo.DirectorURL, graphql.WithHTTPClient(TLSClient))

	req := graphql.NewRequest(`
	mutation ($appID: ID!, $apiSpec: CLOB!, $hostURL String!){
		result: addAPIDefinition(
		  	applicationID: $appID
		  	in: {
				name: "sample-app-api"
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

func (KymaConn *GraphQLConnector) getCertificateSubject() string {
	log.Println("GetCertificateSubject")

	return KymaConn.GraphQLAPIResp.Result.CertificateSigningRequestInfo.Subject
}

func (KymaConn *GraphQLConnector) getEventURL() string {
	log.Println("getEventURL via graphql")

	return KymaConn.GraphQLAPIResp.Result.ManagementPlaneInfo.DirectorURL
}
