package komoot

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (client *Client) Login() (int, error) {

	type loginResponse struct {
		Type  string `json:"type"`
		Error error  `json:"error"`
		Email string `json:"email"`
	}

	if client.Email == "" || client.Password == "" {
		return 0, errors.New("No email or password specified")
	}

	params := url.Values{}
	params.Set("email", client.Email)
	params.Set("password", client.Password)

	req, err := http.NewRequest(http.MethodPost, signInURL, strings.NewReader(params.Encode()))
	if err != nil {
		return 0, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var r loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}

	if r.Type != "logged_in" {
		return 0, errors.New("Invalid email or password")
	}

	req, err = http.NewRequest(http.MethodGet, signInTransferURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err = client.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Check which country / URL we need for Komoot
	bodyStr := string(body)

	prefix := "https://feed-api.komoot.de/v1/"
	suffix := "/feed/"

	start := strings.Index(bodyStr, prefix)
	if start == -1 {
		return 0, errors.New("User ID not found")
	}
	bodyStr = bodyStr[start+len(prefix):]

	end := strings.Index(bodyStr, suffix)
	if end == -1 {
		return 0, errors.New("User ID not found")
	}
	bodyStr = bodyStr[:end]

	return strconv.Atoi(bodyStr)

}
