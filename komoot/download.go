package komoot

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/matryer/try"
)

func (client *Client) Download(tourID int) (string, error) {

	var gpx string

	retryCount := 5

	err := try.Do(func(attempt int) (bool, error) {

		if attempt > 1 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		downloadURL := fmt.Sprintf("https://www.komoot.nl/api/v007/tours/%d.gpx", tourID)

		req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
		if err != nil {
			return attempt < retryCount, err
		}

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return attempt < retryCount, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return attempt < retryCount, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return attempt < retryCount, err
		}

		gpx = string(body)

		return attempt < retryCount, nil

	})
	if err != nil {
		return "", err
	}

	return gpx, nil

}
