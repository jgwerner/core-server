package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	apiclient "github.com/3Blades/go-sdk/client"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

type Args struct {
	ApiKey      string
	ApiRoot     string
	ResourceDir string
	ServerType  string
	Version     string
	Namespace   string
	ProjectID   string
	ServerID    string
	KernelName  string
	Script      string
	Function    string
}

type APIClient struct {
	*apiclient.Threeblades
	AuthInfo runtime.ClientAuthInfoWriterFunc
}

func NewAPIClient(apiRoot, token string) *APIClient {
	transport := httptransport.New(apiRoot, "", []string{"http"})
	cli := apiclient.New(transport, strfmt.Default)
	authInfo := CreateAuthInfo(args.ApiKey)
	return &APIClient{cli, authInfo}
}

func CreateAuthInfo(token string) runtime.ClientAuthInfoWriterFunc {
	return runtime.ClientAuthInfoWriterFunc(func(req runtime.ClientRequest, reg strfmt.Registry) error {
		return req.SetHeaderParam("AUTHORIZATION", fmt.Sprintf("Bearer %s", token))
	})
}

func validateJSON(s []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

func checkToken(apiRoot, token string) bool {
	if token == "" {
		return false
	}
	cli := NewAPIClient(apiRoot, token)
	params := projects.NewProjectsServersAuthParams()
	params.SetNamespace(args.Namespace)
	params.SetProjectID(args.ProjectID)
	params.SetID(args.ServerID)
	authInfo := CreateAuthInfo(token)
	_, err := cli.Projects.ProjectsServersAuth(params, authInfo)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func getTokenFromHeader(header string) (string, error) {
	ok := strings.Contains(header, "Bearer")
	if !ok {
		return "", errors.New("No token")
	}
	return strings.Split(header, " ")[1], nil
}

func getRunner(serverType string) Runner {
	switch serverType {
	case "restful":
		return &RunHTTP{}
	case "proxy":
		return &RunProxy{&RunGeneric{}}
	case "cron":
		return &RunCode{}
	default:
		return &RunGeneric{}
	}
}
