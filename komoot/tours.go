package komoot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (client *Client) Tours(filter string, tourType string) ([]Tour, []byte, error) {
	params := url.Values{}
	params.Set("limit", "5000")
	params.Set("sport_types", tourType)
	params.Set("status", "private")
	params.Set("name", filter)

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://www%s/api/v007/users/%d/tours/?%s", client.komootDomain, client.UserID, params.Encode()),
		nil,
	)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", acceptJson)
	req.Header.Add("Authorization", "Basic "+client.basicAuth())

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var r ToursResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, nil, err
	}

	return r.Embedded.Tours, body, nil
}

func (client *Client) Download(tour Tour) ([]byte, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://www%s/api/v007/tours/%d.gpx", client.komootDomain, tour.ID),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", acceptJson)
	req.Header.Add("Authorization", "Basic "+client.basicAuth())

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (client *Client) basicAuth() string {
	auth := client.Email + ":" + client.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
