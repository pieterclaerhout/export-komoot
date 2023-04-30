package komoot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/danwakefield/fnmatch"
	"github.com/pieterclaerhout/go-log"
)

func (client *Client) Tours(userID int, filter string, tourType string) ([]Tour, []byte, error) {

	params := url.Values{}
	params.Set("limit", "1500")
	params.Set("type", tourType)
	params.Set("status", "private")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://www%s/api/v007/users/%d/tours/?%s", client.komootDomain, userID, params.Encode()), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", acceptJson)

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

	if filter != "" {
		log.Warn(filter)
		filteredTours := []Tour{}

		for _, tour := range r.Embedded.Tours {
			if !fnmatch.Match(filter, tour.Name, 0) {
				continue
			}
			filteredTours = append(filteredTours, tour)
		}

		r.Embedded.Tours = filteredTours

	}

	return r.Embedded.Tours, body, nil
}
