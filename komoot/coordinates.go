package komoot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (client *Client) Coordinates(tour Tour) (*CoordinatesResponse, error) {

	downloadURL := fmt.Sprintf("https://www%s/api/v007/tours/%d/coordinates", client.komootDomain, tour.ID)

	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+client.basicAuth())

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r CoordinatesResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	r.Tour = &tour

	return &r, nil

}
