package komoot

import (
	"fmt"
	"time"
)

type UploadedTour struct {
	Type         string    `json:"type"`
	Source       string    `json:"source"`
	Sport        string    `json:"sport"`
	Constitution int64     `json:"constitution"`
	Name         string    `json:"name"`
	Date         time.Time `json:"date"`
	Embedded     struct {
		Coordinates struct {
			Items []Coordinate `json:"items"`
		} `json:"coordinates"`
	} `json:"_embedded"`
}

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
	T   int64   `json:"t"`
}

func (c Coordinate) Time(startTime time.Time) string {
	t := time.Unix(c.T+startTime.Unix(), 0)
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

type Difficulty struct {
	ExplanationFitness   string `json:"explanation_fitness"`
	ExplanationTechnical string `json:"explanation_technical"`
	Grade                string `json:"grade"`
}

type Path struct {
	Index     int64      `json:"index"`
	Reference string     `json:"reference,omitempty"`
	Location  Coordinate `json:"location"`
}

type Segment struct {
	From int64  `json:"from"`
	To   int64  `json:"to"`
	Type string `json:"type,omitempty"`
}

type Direction struct {
	CardinalDirection string `json:"cardinal_direction"`
	ChangeWay         bool   `json:"change_way"`
	Complex           bool   `json:"complex"`
	Distance          int64  `json:"distance"`
	Index             int64  `json:"index"`
	LastSimilar       int64  `json:"last_similar"`
	StreetName        string `json:"street_name"`
	Type              string `json:"type"`
	WayType           string `json:"way_type"`
}

type Surface struct {
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}

type EmbeddedSurface struct {
	From    int64  `json:"from"`
	To      int64  `json:"to"`
	Element string `json:"element"`
}

type WayType struct {
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}

type EmbeddedWayType struct {
	From    int64  `json:"from"`
	To      int64  `json:"to"`
	Element string `json:"element"`
}

type TourInformation struct {
	Type     string    `json:"type"`
	Segments []Segment `json:"segments"`
}

type Tour struct {
	ID            int64     `json:"id"`
	Type          string    `json:"type"`
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

func (tour *Tour) Filename(ext string) string {
	return fmt.Sprintf(
		"%s_%s_%d_%d.%s",
		tour.Date.Format("2006-01-02"),
		tour.FormattedSport(),
		tour.ID,
		tour.ChangedAt.Unix(),
		ext,
	)
}

func (tour *Tour) FormattedSport() string {
	switch tour.Sport {
	case "mtb":
		return "mountainbike"
	case "racebike":
		return "racebike"
	case "touringbicycle":
		return "touring"
	case "mtb_easy":
		return "gravel"
	case "jogging":
		return "running"
	default:
		return tour.Sport
	}
}
