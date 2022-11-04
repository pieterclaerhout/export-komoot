package komoot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/shabbyrobe/xmlwriter"
	"github.com/tormoder/fit"
)

type loginResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
	Email string `json:"email"`
}

type CoordinatesResponse struct {
	Tour  *Tour        `json:"-"`
	Items []Coordinate `json:"items"`
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
		if r.Tour.Type != "tour_planned" {
			w.StartElem(xmlwriter.Elem{Name: "time"})
			w.WriteText(point.Time(r.Tour.Date))
			w.EndElem("time")
		}
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

	course, err := f.Course()
	if err != nil {
		return nil, err
	}

	course.Course = fit.NewCourseMsg()
	course.Course.Name = fmt.Sprintf("%d %s", r.Tour.ID, r.Tour.Name)

	course.Events = []*fit.EventMsg{}

	firstPoint := r.Items[0]
	lastPoint := r.Items[len(r.Items)-1]

	lap := fit.NewLapMsg()
	lap.TotalDistance = uint32(r.Tour.Distance * 100)
	lap.StartTime = f.FileId.TimeCreated
	lap.Timestamp = f.FileId.TimeCreated
	lap.StartPositionLat = fit.NewLatitudeDegrees(firstPoint.Lat)
	lap.StartPositionLong = fit.NewLongitudeDegrees(firstPoint.Lng)
	lap.EndPositionLat = fit.NewLatitudeDegrees(lastPoint.Lat)
	lap.EndPositionLong = fit.NewLongitudeDegrees(lastPoint.Lng)
	lap.TotalTimerTime = uint32(r.Tour.Duration * 1000)
	lap.TotalAscent = uint16(r.Tour.ElevationUp)
	lap.TotalDescent = uint16(r.Tour.ElevationDown)
	course.Laps = append(course.Laps, lap)

	for i, point := range r.Items {

		if i == 0 {
			ev := fit.NewEventMsg()
			ev.Timestamp = time.Unix(point.T/1000.0, 0)
			ev.Event = fit.EventTimer
			ev.EventType = fit.EventTypeStart
			ev.EventGroup = 0
			course.Events = append(course.Events, ev)
		}

		rec := fit.NewRecordMsg()
		rec.Timestamp = time.Unix(point.T/1000.0, 0)
		rec.PositionLat = fit.NewLatitudeDegrees(point.Lat)
		rec.PositionLong = fit.NewLongitudeDegrees(point.Lng)
		rec.Altitude = uint16((point.Alt + 500.0) * 5.0)
		course.Records = append(course.Records, rec)

		if i == len(r.Items)-1 {
			ev := fit.NewEventMsg()
			ev.Timestamp = time.Unix(point.T/1000.0, 0)
			ev.Event = fit.EventTimer
			ev.EventType = fit.EventTypeStopAll
			ev.EventGroup = 0
			course.Events = append(course.Events, ev)
		}

	}

	if err = fit.Encode(out, f, binary.LittleEndian); err != nil {
		return nil, err
	}

	return out.Bytes(), nil

}

type ToursResponse struct {
	Embedded struct {
		Tours []Tour `json:"tours"`
	} `json:"_embedded"`
}

type UploadTourResponse struct {
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Embedded struct {
		Items   []UploadedTour `json:"items"`
		Matched MatchedTour    `json:"matched"`
	} `json:"_embedded"`
	Message string `json:"message"`
}

type MatchedTour struct {
	Constitution  int64      `json:"constitution"`
	Status        string     `json:"status"`
	Date          string     `json:"date"`
	Difficulty    Difficulty `json:"difficulty"`
	Distance      float64    `json:"distance"`
	Duration      float64    `json:"duration"`
	ElevationDown float64    `json:"elevation_down"`
	ElevationUp   float64    `json:"elevation_up"`
	Name          string     `json:"name"`
	Path          []Path     `json:"path"`
	Query         string     `json:"query"`
	Segments      []Segment  `json:"segments"`
	Source        string     `json:"source"`
	Sport         string     `json:"sport"`
	Summary       struct {
		Surfaces []Surface `json:"surfaces"`
		WayTypes []WayType `json:"way_types"`
	} `json:"summary"`
	TourInformation []TourInformation `json:"tour_information"`
	Type            string            `json:"type"`
	Embedded        struct {
		Coordinates struct {
			Items []Coordinate `json:"items"`
		} `json:"coordinates"`
		Directions struct {
			Items []Direction `json:"items"`
		} `json:"directions"`
		Surfaces struct {
			Items []EmbeddedSurface `json:"items"`
		} `json:"surfaces"`
		WayTypes struct {
			Items []EmbeddedWayType `json:"items"`
		} `json:"way_types"`
	} `json:"_embedded"`
}
