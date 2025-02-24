package komoot

import (
	"net/http"
)

// See: https://github.com/janthomas89/komoot-api-client/blob/master/src/KomootApiClient.php
// See: https://static.komoot.de/doc/auth/oauth2.html

type Client struct {
	Email        string
	Password     string
	UserID       int64
	IsLoggedIn   bool
	komootDomain string
	httpClient   *http.Client
}

func NewClient(email string, password string, userID int64) *Client {
	return &Client{
		Email:        email,
		Password:     password,
		UserID:       userID,
		IsLoggedIn:   false,
		komootDomain: ".komoot.com",
		httpClient: &http.Client{
			CheckRedirect: nil,
		},
	}
}
