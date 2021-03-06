package komoot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type ToursResponse struct {
	Embedded struct {
		Tours []Tour `json:"tours"`
	} `json:"_embedded"`
}

type Tour struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Sport     string    `json:"sport"`
	Status    string    `json:"status"`
	Date      time.Time `json:"date"`
	ChangedAt time.Time `json:"changed_at"`
}

func (tour Tour) Filename() string {
	return fmt.Sprintf("%d_%d.gpx", tour.ID, tour.ChangedAt.Unix())
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

	// https://www.komoot.nl/api/v007/users/471950076586/tours/?limit=24&sport_types=racebike%2Ce_racebike&type=tour_planned&sort_field=date&sort_direction=desc&name=&status=private&hl=nl

	params := url.Values{}
	params.Set("limit", "1000")
	params.Set("type", "tour_planned")
	params.Set("status", "private")
	params.Set("name", filter)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://www.komoot.nl/api/v007/users/%d/tours/?%s", userID, params.Encode()), nil)
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

	return r.Embedded.Tours, body, nil

}
