package komoot

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pieterclaerhout/go-log"
)

func (client *Client) Login() (int, error) {

	if client.Email == "" || client.Password == "" {
		return 0, errors.New("no email or password specified")
	}

	params := url.Values{}
	params.Set("email", client.Email)
	params.Set("password", client.Password)

	req, err := http.NewRequest(http.MethodPost, loginUrl, strings.NewReader(params.Encode()))
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

	if r.Error == "rate_limit" {
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		log.Warn("Rate limit, retrying after", retryAfter, "seconds")
		time.Sleep(time.Duration(retryAfter) * time.Second)
		return client.Login()
	}

	if r.Type != "logged_in" {
		return 0, errors.New("login failed: " + r.Error)
	}

	req, err = http.NewRequest(http.MethodGet, transferUrl, nil)
	if err != nil {
		return 0, err
	}

	resp, err = client.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	for _, cookie := range resp.Cookies() {
		if strings.HasPrefix(cookie.Domain, ".komoot") {
			client.komootDomain = cookie.Domain
			break
		}
	}

	return client.getUserIdFromBody(body)

}

func (client *Client) getUserIdFromBody(body []byte) (int, error) {

	bodyStr := string(body)

	prefix := "https://feed-api.komoot.de/v1/"
	suffix := "/feed/"

	start := strings.Index(bodyStr, prefix)
	if start == -1 {
		return 0, errors.New("user ID not found")
	}
	bodyStr = bodyStr[start+len(prefix):]

	end := strings.Index(bodyStr, suffix)
	if end == -1 {
		return 0, errors.New("user ID not found")
	}
	bodyStr = bodyStr[:end]

	return strconv.Atoi(bodyStr)

}
