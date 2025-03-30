package komoot

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CoordinatesResponse struct {
	Tour  *Tour        `json:"-"`
	Items []Coordinate `json:"items"`
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

type RoutePlanResponse struct {
	Duration float64 `json:"duration"`
}
