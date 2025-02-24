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

func (client *Client) Upload(userID int64, name string, gpxData string, sport string, makeRoundtrip bool, overwrite bool) (*MatchedTour, error) {
	if overwrite {
		existingTours, _, err := client.Tours(userID, name, "")
		if err != nil {
			return nil, err
		}

		if len(existingTours) == 1 {
			log.Warn("Overwriting existing tour:", name)
			client.deleteTour(existingTours[0])
		}
	}

	importedGpx, err := client.importGpx(gpxData)
	if err != nil {
		return nil, err
	}

	tour, err := client.importTour(importedGpx, sport)
	if err != nil {
		return nil, err
	}

	duration, err := client.planRoute(tour)
	if err != nil {
		return nil, err
	}

	tour = client.updateTourSettings(tour, name, sport, makeRoundtrip, duration)

	if err := client.createTour(tour); err != nil {
		return nil, err
	}
	return tour, nil
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
	req.Header.Add("Authorization", "Basic "+client.basicAuth())

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
	req.Header.Add("Authorization", "Basic "+client.basicAuth())

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
	log.DebugDump(string(body), "body:")

	var r UploadTourResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	return &r.Embedded.Matched, nil
}

func (client *Client) updateTourSettings(tour *MatchedTour, name string, sport string, makeRoundtrip bool, duration float64) *MatchedTour {
	tour.Name = name
	tour.Sport = sport
	tour.Status = "public"
	tour.Type = "tour_planned"
	tour.Constitution = 4
	tour.Duration = duration

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
		fmt.Sprintf("https://www%s/api/v007/tours/?reroute=true&hl=nl", client.komootDomain),
		bytes.NewBuffer(tourJson),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeJson)
	req.Header.Set("Accept", acceptJson)
	req.Header.Add("Authorization", "Basic "+client.basicAuth())

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
	log.DebugDump(string(body), "body:")

	return nil
}

func (client *Client) planRoute(tour *MatchedTour) (float64, error) {
	paths := []RoutePlanRequestPath{}
	for _, path := range tour.Path {
		paths = append(paths, RoutePlanRequestPath{
			Reference: path.Reference,
			Location:  path.Location,
		})
	}

	segments := []RoutePlanRequestSegment{}
	for _, segment := range tour.Segments {
		segments = append(segments, RoutePlanRequestSegment{
			Type: segment.Type,
		})
	}

	bodyJson, err := json.Marshal(RoutePlanRequest{
		Constitution: 4,
		Path:         paths,
		Segments:     segments,
		Sport:        tour.Sport,
	})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(`https://www%s/api/routing/tour?sport=racebike&_embedded=coordinates%%2Cway_types%%2Csurfaces%%2Cdirections`, client.komootDomain),
		bytes.NewBuffer(bodyJson),
	)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", contentTypeJson)
	req.Header.Set("Accept", acceptJson)
	req.Header.Add("Authorization", "Basic "+client.basicAuth())

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	log.DebugSeparator("plan tour")
	log.Debug(string(body))

	var r RoutePlanResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return 0, err
	}

	return r.Duration, nil
}

func (client *Client) deleteTour(tour Tour) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("https://www%s/api/v007/tours/%d?hl=nl", client.komootDomain, tour.ID),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeJson)
	req.Header.Set("Accept", acceptJson)
	req.Header.Add("Authorization", "Basic "+client.basicAuth())

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.DebugSeparator("delete tour")
	log.DebugDump(string(body), "body:")

	return nil

}
