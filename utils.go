package main

import (
	"encoding/json"
	"errors"
	"strings"

	apiclient "github.com/3Blades/go-sdk/client"
	"github.com/3Blades/go-sdk/client/auth"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

type Args struct {
	ApiKey      string
	ApiRoot     string
	ResourceDir string
	ServerType  string
	Namespace   string
	ProjectID   string
	ServerID    string
	KernelName  string
	Script      string
	Function    string
}

func APIClient(apiRoot, token string) *apiclient.Threeblades {
	transport := httptransport.New(apiRoot, "", []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("AUTHORIZATION", "header",
		"Token "+token)
	return apiclient.New(transport, strfmt.Default)
}

func validateJSON(s []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

func checkToken(apiRoot, tokenHeader string) bool {
	if tokenHeader == "" {
		return false
	}
	token, err := getTokenFromHeader(tokenHeader)
	if err != nil {
		logger.Printf("Error getting token from header: %s", err.Error())
		return false
	}
	cli := APIClient(apiRoot, token)
	params := auth.NewAuthJwtTokenVerifyCreateParams()
	_, err = cli.Auth.AuthJwtTokenVerifyCreate(params)
	if err != nil {
		return false
	}
	return true
}

func getTokenFromHeader(header string) (string, error) {
	ok := strings.Contains(header, "Token")
	if !ok {
		return "", errors.New("No token")
	}
	return header[len(header)-40:], nil
}

func getRunner(serverType string) Runner {
	switch serverType {
	case "restful":
		return &RunHTTP{}
	case "cron":
		return &RunCode{}
	default:
		return &RunGeneric{}
	}
}
