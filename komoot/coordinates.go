package komoot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shabbyrobe/xmlwriter"
)

type CoordinatesResponse struct {
	Items []Coordinate `json:"items"`
}

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
	T   int64   `json:"t"`
}

func (c Coordinate) Time() string {
	t := time.Unix(c.T/1000.0, 0)
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func (r CoordinatesResponse) GPX(name string) []byte {

	out := &bytes.Buffer{}

	indenter := xmlwriter.NewStandardIndenter()
	indenter.IndentString = "\t"

	w := xmlwriter.Open(out)
	w.Indenter = indenter

	w.StartDoc(xmlwriter.Doc{})
	w.StartElem(xmlwriter.Elem{
		Name: "gpx",
		Attrs: []xmlwriter.Attr{
			{Name: "creator", Value: "export-komoot"},
			{Name: "version", Value: "1.1"},
			{Name: "xmlns", Value: "http://www.topografix.com/GPX/1/1"},
			{Name: "xmlns:xsi", Value: "http://www.w3.org/2001/XMLSchema-instance"},
			{Name: "xsi:schemaLocation", Value: "http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd"},
		},
	})
	w.StartElem(xmlwriter.Elem{Name: "metadata"})
	w.StartElem(xmlwriter.Elem{Name: "name"})
	w.WriteText(name)
	w.EndElem("name")
	w.EndElem("metadata")
	w.StartElem(xmlwriter.Elem{Name: "trk"})
	w.StartElem(xmlwriter.Elem{Name: "name"})
	w.WriteText(name)
	w.EndElem("name")
	w.StartElem(xmlwriter.Elem{Name: "trkseg"})
	for _, point := range r.Items {
		w.StartElem(xmlwriter.Elem{Name: "trkpt"})
		w.WriteAttr(xmlwriter.Attr{Name: "lat"}.Float64(point.Lat))
		w.WriteAttr(xmlwriter.Attr{Name: "lon"}.Float64(point.Lng))
		w.StartElem(xmlwriter.Elem{Name: "ele"})
		w.WriteText(fmt.Sprintf("%f", point.Alt))
		w.EndElem("ele")
		w.StartElem(xmlwriter.Elem{Name: "time"})
		w.WriteText(point.Time())
		w.EndElem("time")
		w.EndElem("trkpt")
	}
	w.EndElem("trkseg")
	w.EndElem("trk")
	w.EndElem("gpx")
	w.EndAllFlush()

	return out.Bytes()

}

func (client *Client) Coordinates(tour Tour) (*CoordinatesResponse, error) {

	downloadURL := fmt.Sprintf("https://www.komoot.nl/api/v007/tours/%d/coordinates", tour.ID)

	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r CoordinatesResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	return &r, nil

}
