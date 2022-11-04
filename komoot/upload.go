package komoot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (client *Client) Upload(name string, gpxData string, sport string) error {
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

	var r UploadTourResponse
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

	var r2 UploadTourResponse
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

	tour2.Segments = append(tour2.Segments, Segment{
		From: lastPoint.Index,
		To:   lastPoint.Index + 1,
		Type: "Routed",
	})

	lastCoordinate := tour2.Embedded.Coordinates.Items[len(tour2.Embedded.Coordinates.Items)-1]

	tour2.Embedded.Coordinates.Items = append(tour2.Embedded.Coordinates.Items, Coordinate{
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

	return nil
}
