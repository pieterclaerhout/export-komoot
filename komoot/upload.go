package komoot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pieterclaerhout/go-log"
)

type ImportedToursResponse struct {
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Embedded struct {
		Items   []ImportedTour `json:"items"`
		Matched struct {
			Constitution int64  `json:"constitution"`
			Status       string `json:"status"`
			Date         string `json:"date"`
			Difficulty   struct {
				ExplanationFitness   string `json:"explanation_fitness"`
				ExplanationTechnical string `json:"explanation_technical"`
				Grade                string `json:"grade"`
			} `json:"difficulty"`
			Distance      float64 `json:"distance"`
			Duration      float64 `json:"duration"`
			ElevationDown float64 `json:"elevation_down"`
			ElevationUp   float64 `json:"elevation_up"`
			Name          string  `json:"name"`
			Path          []struct {
				Index     int64  `json:"index"`
				Reference string `json:"reference,omitempty"`
				Location  struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
					Alt float64 `json:"alt"`
				} `json:"location"`
			} `json:"path"`
			Query    string `json:"query"`
			Segments []struct {
				From int64  `json:"from"`
				To   int64  `json:"to"`
				Type string `json:"type"`
			} `json:"segments"`
			Source  string `json:"source"`
			Sport   string `json:"sport"`
			Summary struct {
				Surfaces []struct {
					Amount float64 `json:"amount"`
					Type   string  `json:"type"`
				} `json:"surfaces"`
				WayTypes []struct {
					Amount float64 `json:"amount"`
					Type   string  `json:"type"`
				} `json:"way_types"`
			} `json:"summary"`
			TourInformation []struct {
				Type     string `json:"type"`
				Segments []struct {
					From int64 `json:"from"`
					To   int64 `json:"to"`
				} `json:"segments"`
			} `json:"tour_information"`
			Type     string `json:"type"`
			Embedded struct {
				Coordinates struct {
					Items []struct {
						Lat float64 `json:"lat"`
						Lng float64 `json:"lng"`
						Alt float64 `json:"alt"`
						T   float64 `json:"t"`
					} `json:"items"`
				} `json:"coordinates"`
				Directions struct {
					Items []struct {
						CardinalDirection string `json:"cardinal_direction"`
						ChangeWay         bool   `json:"change_way"`
						Complex           bool   `json:"complex"`
						Distance          int64  `json:"distance"`
						Index             int64  `json:"index"`
						LastSimilar       int64  `json:"last_similar"`
						StreetName        string `json:"street_name"`
						Type              string `json:"type"`
						WayType           string `json:"way_type"`
					} `json:"items"`
				} `json:"directions"`
				Surfaces struct {
					Items []struct {
						From    int64  `json:"from"`
						To      int64  `json:"to"`
						Element string `json:"element"`
					} `json:"items"`
				} `json:"surfaces"`
				WayTypes struct {
					Items []struct {
						From    int64  `json:"from"`
						To      int64  `json:"to"`
						Element string `json:"element"`
					} `json:"items"`
				} `json:"way_types"`
			} `json:"_embedded"`
		} `json:"matched"`
	} `json:"_embedded"`
	Message string `json:"message"`
}

type ImportedTour struct {
	Type         string    `json:"type"`
	Source       string    `json:"source"`
	Sport        string    `json:"sport"`
	Constitution int64     `json:"constitution"`
	Name         string    `json:"name"`
	Date         time.Time `json:"date"`
	Embedded     struct {
		Coordinates struct {
			Items []ImportedCoordinate `json:"items"`
		} `json:"coordinates"`
	} `json:"_embedded"`
}

type ImportedCoordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}

func (client *Client) Upload(name string, gpxData string, sport string) error {
	log.Warn("Importing GPX")

	params := url.Values{}
	params.Set("data_type", "gpx")

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://www%s/api/routing/import/files/?%s", client.komootDomain, params.Encode()), bytes.NewBuffer([]byte(gpxData)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/hal+json,application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r ImportedToursResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return err
	}

	if len(r.Embedded.Items) != 1 {
		return errors.New("import failed: " + r.Message)
	}

	tourJson, err := json.Marshal(r.Embedded.Items[0])
	if err != nil {
		return err
	}

	req, err = http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://www%s/api/routing/import/tour?sport=%s&_embedded=way_types%%2Csurfaces%%2Cdirections%%2Ccoordinates", client.komootDomain, sport),
		bytes.NewBuffer(tourJson),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/hal+json,application/json")

	resp, err = client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r2 ImportedToursResponse
	if err := json.Unmarshal(body, &r2); err != nil {
		return err
	}

	tour2 := &r2.Embedded.Matched
	tour2.Name = name
	tour2.Sport = sport
	tour2.Status = "public"
	tour2.Type = "tour_planned"
	tour2.Constitution = 4

	firstPoint := tour2.Path[0]
	lastPoint := tour2.Path[len(tour2.Path)-1]
	firstPoint.Reference = "special:back"
	firstPoint.Index = lastPoint.Index + 1

	tour2.Path = append(tour2.Path, firstPoint)

	tour2.Segments = append(tour2.Segments, struct {
		From int64  "json:\"from\""
		To   int64  "json:\"to\""
		Type string "json:\"type\""
	}{
		From: lastPoint.Index,
		To:   lastPoint.Index + 1,
		Type: "Routed",
	})

	lastCoordinate := tour2.Embedded.Coordinates.Items[len(tour2.Embedded.Coordinates.Items)-1]

	tour2.Embedded.Coordinates.Items = append(tour2.Embedded.Coordinates.Items, struct {
		Lat float64 "json:\"lat\""
		Lng float64 "json:\"lng\""
		Alt float64 "json:\"alt\""
		T   float64 "json:\"t\""
	}{
		Lat: firstPoint.Location.Lat,
		Lng: firstPoint.Location.Lng,
		Alt: firstPoint.Location.Alt,
		T:   lastCoordinate.T + 1,
	})

	tourJson2, err := json.Marshal(tour2)
	if err != nil {
		return err
	}

	req, err = http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://www%s/api/v007/tours/?hl=nl", client.komootDomain),
		bytes.NewBuffer(tourJson2),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/hal+json,application/json")

	resp, err = client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// body, err = io.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }

	return nil
}
