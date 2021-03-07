package komoot

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shabbyrobe/xmlwriter"
	"github.com/tormoder/fit"
)

type CoordinatesResponse struct {
	Tour  *Tour        `json:"-"`
	Items []Coordinate `json:"items"`
}

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
	T   int64   `json:"t"`
}

func (c Coordinate) Time() string {
	// What about nanoseconds?
	t := time.Unix(c.T/1000.0, 0)
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func (r CoordinatesResponse) GPX() []byte {

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
	w.WriteText(r.Tour.Name)
	w.EndElem("name")
	w.EndElem("metadata")
	w.StartElem(xmlwriter.Elem{Name: "trk"})
	w.StartElem(xmlwriter.Elem{Name: "name"})
	w.WriteText(r.Tour.Name)
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

func (r CoordinatesResponse) Fit() ([]byte, error) {

	out := &bytes.Buffer{}

	hdr := fit.NewHeader(fit.V20, true)
	f, err := fit.NewFile(fit.FileTypeCourse, hdr)
	if err != nil {
		return nil, err
	}

	f.FileId.TimeCreated = time.Now()
	f.FileId.SerialNumber = uint32(time.Now().Unix())
	// f.FileId.Manufacturer = fit.ManufacturerGarmin
	// f.FileId.Product = uint16(fit.GarminProductConnect)
	// f.FileId.ProductName = "export-komoot"
	// f.FileId.Number = 1

	// f.FileCreator = fit.NewFileCreatorMsg()
	// f.FileCreator.SoftwareVersion = 950

	act, err := f.Activity()
	if err != nil {
		return nil, err
	}

	act.Events = []*fit.EventMsg{}

	course, err := f.Course()
	if err != nil {
		return nil, err
	}

	course.Course = fit.NewCourseMsg()
	// course.Course.Capabilities = fit.CourseCapabilities(771)
	// course.Course.Sport = fit.SportCycling
	course.Course.Name = fmt.Sprintf("___%d %s", r.Tour.ID, r.Tour.Name)

	firstPoint := r.Items[0]
	lastPoint := r.Items[len(r.Items)-1]

	lap := fit.NewLapMsg()
	// lap.MessageIndex = fit.MessageIndex(0)
	// lap.EventType = fit.EventTypeStop
	// lap.Sport = fit.SportGeneric
	// lap.Event = fit.EventLap
	// lap.LapTrigger = fit.LapTriggerSessionEnd
	lap.TotalDistance = uint32(2696800) // To calculate
	lap.StartTime = f.FileId.TimeCreated
	lap.Timestamp = f.FileId.TimeCreated
	lap.StartPositionLat = fit.NewLatitudeDegrees(firstPoint.Lat)
	lap.StartPositionLong = fit.NewLongitudeDegrees(firstPoint.Lng)
	lap.EndPositionLat = fit.NewLatitudeDegrees(lastPoint.Lat)
	lap.EndPositionLong = fit.NewLongitudeDegrees(lastPoint.Lng)
	lap.TotalTimerTime = uint32(lastPoint.T)
	// lap.TotalMovingTime = uint32(lastPoint.T)
	// lap.AvgSpeed = 0
	course.Laps = append(course.Laps, lap)

	for i, point := range r.Items {

		if i == 0 {
			ev := fit.NewEventMsg()
			ev.Timestamp = time.Unix(point.T/1000.0, 0)
			ev.Event = fit.EventTimer
			ev.EventType = fit.EventTypeStart
			ev.EventGroup = 0
			// act.Events = append(act.Events, ev)
		}

		rec := fit.NewRecordMsg()
		rec.Timestamp = time.Unix(point.T/1000.0, 0)
		// r.Timestamp = f.FileId.TimeCreated.Add(time.Duration(point.T/1000.0) * time.Second)
		rec.PositionLat = fit.NewLatitudeDegrees(point.Lat)
		rec.PositionLong = fit.NewLongitudeDegrees(point.Lng)
		rec.Altitude = uint16((point.Alt + 500.0) * 5.0)
		rec.Distance = uint32(i * 1000)
		course.Records = append(course.Records, rec)

		if i == len(r.Items)-1 {
			ev := fit.NewEventMsg()
			ev.Timestamp = time.Unix(point.T/1000.0, 0)
			ev.Event = fit.EventTimer
			ev.EventType = fit.EventTypeStopAll
			ev.EventGroup = 0
			// act.Events = append(act.Events, ev)
		}

	}

	if err = fit.Encode(out, f, binary.LittleEndian); err != nil {
		return nil, err
	}

	return out.Bytes(), nil

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

	r.Tour = &tour

	return &r, nil

}
