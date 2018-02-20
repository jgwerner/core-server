package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	apiclient "github.com/IllumiDesk/go-sdk/client"
	"github.com/IllumiDesk/go-sdk/client/projects"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	cache "github.com/patrickmn/go-cache"
)

var store = cache.New(5*time.Second, time.Minute)

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
	SecretKey   string
}

type APIClient struct {
	*apiclient.Illumidesk
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
	_, found := store.Get(token)
	if found {
		return true
	}
	cli := NewAPIClient(apiRoot, token)
	params := projects.NewProjectsServersAuthParams()
	params.SetNamespace(args.Namespace)
	params.SetProject(args.ProjectID)
	params.SetServer(args.ServerID)
	authInfo := CreateAuthInfo(token)
	_, err := cli.Projects.ProjectsServersAuth(params, authInfo)
	if err != nil {
		log.Println(err)
		return false
	}
	store.Set(token, true, cache.DefaultExpiration)
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
