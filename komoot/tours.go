package komoot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/danwakefield/fnmatch"
	"github.com/pieterclaerhout/go-log"
)

type ToursResponse struct {
	Embedded struct {
		Tours []Tour `json:"tours"`
	} `json:"_embedded"`
}

type Tour struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Sport         string    `json:"sport"`
	Status        string    `json:"status"`
	Date          time.Time `json:"date"`
	Distance      float64   `json:"distance"`
	Duration      int64     `json:"duration"`
	ElevationUp   float64   `json:"elevation_up"`
	ElevationDown float64   `json:"elevation_down"`
	ChangedAt     time.Time `json:"changed_at"`
}

func (tour Tour) Filename(ext string) string {
	return fmt.Sprintf("%d_%d.%s", tour.ID, tour.ChangedAt.Unix(), ext)
}

func (tour Tour) IsCycling() bool {
	switch tour.Sport {
	case "mtb", "racebike", "touringbicycle", "mtb_easy":
		return true
	default:
		return false
	}
}
func (tour Tour) FormattedSport() string {
	switch tour.Sport {
	case "mtb":
		return "mountainbike"
	case "racebike":
		return "racebike"
	case "touringbicycle":
		return "touring"
	case "mtb_easy":
		return "gravel"
	default:
		return tour.Sport
	}
}

func (client *Client) Tours(userID int, filter string) ([]Tour, []byte, error) {

	params := url.Values{}
	params.Set("limit", "1000")
	params.Set("type", "tour_planned")
	params.Set("status", "private")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://www%s/api/v007/users/%d/tours/?%s", client.komootDomain, userID, params.Encode()), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/hal+json,application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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
