package komoot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pieterclaerhout/go-log"
)

func (client *Client) Upload(name string, gpxData string, sport string, makeRoundtrip bool) error {
	importedGpx, err := client.importGpx(gpxData)
	if err != nil {
		return err
	}

	tour, err := client.importTour(importedGpx, sport)
	if err != nil {
		return err
	}

	tour = client.updateTourSettings(tour, name, sport, makeRoundtrip)

	return client.createTour(tour)
}

func (client *Client) importGpx(gpxData string) (*UploadedTour, error) {
	params := url.Values{}
	params.Set("data_type", "gpx")

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://www%s/api/routing/import/files/?%s", client.komootDomain, params.Encode()), bytes.NewBuffer([]byte(gpxData)),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", acceptJson)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.DebugSeparator("import gpx")
	log.Debug(string(body))

	var r UploadTourResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	if len(r.Embedded.Items) != 1 {
		return nil, errors.New("import failed: " + r.Message)
	}

	return &r.Embedded.Items[0], nil
}

func (client *Client) importTour(tour *UploadedTour, sport string) (*MatchedTour, error) {
	tourJson, err := json.Marshal(tour)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://www%s/api/routing/import/tour?sport=%s&_embedded=way_types%%2Csurfaces%%2Cdirections%%2Ccoordinates", client.komootDomain, sport),
		bytes.NewBuffer(tourJson),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentTypeJson)
	req.Header.Set("Accept", acceptJson)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.DebugSeparator("import tour")
	log.Debug(string(body))

	var r UploadTourResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	return &r.Embedded.Matched, nil
}

func (client *Client) updateTourSettings(tour *MatchedTour, name string, sport string, makeRoundtrip bool) *MatchedTour {
	tour.Name = name
	tour.Sport = sport
	tour.Status = "public"
	tour.Type = "tour_planned"
	tour.Constitution = 4

	if makeRoundtrip {
		firstPoint := tour.Path[0]
		lastPoint := tour.Path[len(tour.Path)-1]
		firstPoint.Reference = "special:back"
		firstPoint.Index = lastPoint.Index + 1

		tour.Path = append(tour.Path, firstPoint)

		tour.Segments = append(tour.Segments, Segment{
			From: lastPoint.Index,
			To:   lastPoint.Index + 1,
			Type: "Routed",
		})

		lastCoordinate := tour.Embedded.Coordinates.Items[len(tour.Embedded.Coordinates.Items)-1]

		tour.Embedded.Coordinates.Items = append(tour.Embedded.Coordinates.Items, Coordinate{
			Lat: firstPoint.Location.Lat,
			Lng: firstPoint.Location.Lng,
			Alt: firstPoint.Location.Alt,
			T:   lastCoordinate.T + 1,
		})
	}

	return tour
}

func (client *Client) createTour(tour *MatchedTour) error {
	tourJson, err := json.Marshal(tour)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://www%s/api/v007/tours/?hl=nl", client.komootDomain),
		bytes.NewBuffer(tourJson),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeJson)
	req.Header.Set("Accept", acceptJson)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.DebugSeparator("create tour")
	log.Debug(string(body))

	return nil
}
