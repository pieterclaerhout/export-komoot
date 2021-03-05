package komoot

import (
	"net/http"
	"net/http/cookiejar"
)

// See: https://github.com/janthomas89/komoot-api-client/blob/master/src/KomootApiClient.php
// See: https://static.komoot.de/doc/auth/oauth2.html

const signInURL = "https://account.komoot.com/v1/signin"
const signInTransferURL = "https://account.komoot.com/actions/transfer?type=signin"

// const toursURLTpl = "https://www.komoot.de/user/%s/tours"
// const tourGpxURLTpl = "https://www.komoot.de/tour/%s/download"

type Client struct {
	Email      string
	Password   string
	IsLoggedIn bool
	cookieJar  *cookiejar.Jar
	httpClient *http.Client
}

func NewClient(email string, password string) *Client {
	cookieJar, _ := cookiejar.New(nil)
	return &Client{
		Email:      email,
		Password:   password,
		IsLoggedIn: false,
		cookieJar:  cookieJar,
		httpClient: &http.Client{
			CheckRedirect: nil,
			Jar:           cookieJar,
		},
	}
}
