package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	keycloakEndpoint  string
	keycloakRealm     string
	clientID          string
	requestedAudience string
	username          string
	password          string
	scope             string
	accessToken       bool
	refreshToken      bool

	realmURL string
)

func init() {
	flag.StringVar(&keycloakEndpoint, "keycloak-endpoint", "http://localhost:6060", "the keycloak base endpoint")
	flag.StringVar(&keycloakRealm, "keycloak-realm", "demo", "the realm to use")
	flag.StringVar(&clientID, "client-id", "api", "the client id for this client")
	flag.StringVar(&username, "username", "", "the username to sign in with")
	flag.StringVar(&password, "password", "", "the password corresponding to the username")
	flag.StringVar(&scope, "scope", "profile email", "the requested scope to get a token with the right audience")
	flag.BoolVar(&accessToken, "access-token", false, "print access token only")
	flag.BoolVar(&refreshToken, "refresh-token", false, "print refresh token only")
}

// direct grant access
func main() {
	flag.Parse()

	realmURL = fmt.Sprintf("%s/auth/realms/%s", keycloakEndpoint, keycloakRealm)
	resp, err := directGrant(realmURL, clientID, username, password)
	if err != nil {
		log.Println(err)
	}

	if accessToken {
		fmt.Println(resp.AccessToken)
		os.Exit(0)
	}

	if refreshToken {
		fmt.Println(resp.RefreshToken)
		os.Exit(0)
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(data))
}

type directGrantResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expired_in"`
	RefreshExpiresIn int    `json:"refresh_expired_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

// for public clients only
func directGrant(realmURL string, clientID string, username string, password string) (*directGrantResponse, error) {
	directGrantURL := fmt.Sprintf("%s/protocol/openid-connect/token", realmURL)
	//log.Printf("direct grant url: %s\n", directGrantURL)

	form := url.Values{
		"grant_type": []string{"password"},
		"username":   []string{username},
		"password":   []string{password},
		"client_id":  []string{clientID},
		"scope":      []string{scope},
	}
	resp, err := http.PostForm(directGrantURL, form)
	if err != nil {
		return nil, fmt.Errorf("failed to issue http request: %s", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve token: %s", string(data))
	}

	var grantResponse directGrantResponse
	err = json.Unmarshal(data, &grantResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err)
	}
	return &grantResponse, nil
}
