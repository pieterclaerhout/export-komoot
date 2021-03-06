package komoot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (client *Client) Coordinates(tour Tour) ([]byte, bool, error) {

	downloadURL := fmt.Sprintf("https://www.komoot.nl/api/v007/tours/%d/coordinates", tour.ID)

	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return []byte(""), false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}

	if body == nil || len(body) == 0 {
		return nil, false, errors.New("empty gpx file")
	}

	return body, true, nil

}
