package komoot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
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

type Path struct {
	Location Location `json:"location"`
	Index    int64    `json:"index"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}

type Tour struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Sport     string    `json:"sport"`
	Status    string    `json:"status"`
	Date      time.Time `json:"date"`
	ChangedAt time.Time `json:"changed_at"`
	Path      []Path    `json:"path"`
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

func (tour Tour) RecreatedGPX() []byte {

	var out bytes.Buffer

	out.WriteString("<?xml version='1.0' encoding='UTF-8'?>\n")
	out.WriteString("<gpx creator=\"export-komoot\" version=\"1.1\" xmlns=\"http://www.topografix.com/GPX/1/1\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xsi:schemaLocation=\"http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd\">")
	out.WriteString("<metadata>")
	out.WriteString("<name>" + html.EscapeString(tour.Name) + "</name>")
	out.WriteString("</metadata>")
	out.WriteString("<trk>")
	out.WriteString("<name>" + html.EscapeString(tour.Name) + "</name>")
	out.WriteString("<trkseg>")
	for _, location := range tour.Path {
		point := location.Location
		out.WriteString(fmt.Sprintf("<trkpt lat=\"%f\" lon=\"%f\">", point.Lat, point.Lng))
		out.WriteString(fmt.Sprintf("<ele>%f</ele>", point.Alt))
		out.WriteString("</trkpt>")
	}
	out.WriteString("</trkseg>")
	out.WriteString("</trk>")
	out.WriteString("</gpx>")

	return out.Bytes()

}

func (client *Client) Tours(userID int) ([]Tour, []byte, error) {

	// https://www.komoot.nl/api/v007/users/471950076586/tours/?limit=24&sport_types=racebike%2Ce_racebike&type=tour_planned&sort_field=date&sort_direction=desc&name=&status=private&hl=nl

	params := url.Values{}
	params.Set("limit", "1000")
	params.Set("type", "tour_planned")
	params.Set("status", "private")
	// params.Set("name", "__")

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
