package komoot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (client *Client) Download(tourID int) ([]byte, error) {

	downloadURL := fmt.Sprintf("https://www.komoot.nl/api/v007/tours/%d.gpx", tourID)

	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if body == nil || len(body) == 0 {
		return nil, errors.New("empty gpx file")
	}

	return body, nil

}
