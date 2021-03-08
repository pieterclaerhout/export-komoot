package komoot

import (
	"net/http"
	"net/http/cookiejar"
)

// See: https://github.com/janthomas89/komoot-api-client/blob/master/src/KomootApiClient.php
// See: https://static.komoot.de/doc/auth/oauth2.html

type Client struct {
	Email        string
	Password     string
	IsLoggedIn   bool
	komootDomain string
	cookieJar    *cookiejar.Jar
	httpClient   *http.Client
}

func NewClient(email string, password string) *Client {
	cookieJar, _ := cookiejar.New(nil)
	return &Client{
		Email:        email,
		Password:     password,
		IsLoggedIn:   false,
		komootDomain: ".komoot.com",
		cookieJar:    cookieJar,
		httpClient: &http.Client{
			CheckRedirect: nil,
			Jar:           cookieJar,
		},
	}
}
