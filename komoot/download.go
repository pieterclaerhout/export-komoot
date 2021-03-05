package komoot

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func (client *Client) Download(tourID int) (string, error) {

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://www.komoot.nl/tour/%d/download", tourID), nil)
	if err != nil {
		return "", err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil

}
