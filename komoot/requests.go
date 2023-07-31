package komoot

type RoutePlanRequestPath struct {
	Reference string     `json:"reference,omitempty"`
	Location  Coordinate `json:"location"`
}

type RoutePlanRequestSegment struct {
	Type string `json:"type,omitempty"`
}

type RoutePlanRequest struct {
	Constitution int64                     `json:"constitution"`
	Path         []RoutePlanRequestPath    `json:"path"`
	Segments     []RoutePlanRequestSegment `json:"segments"`
	Sport        string                    `json:"sport"`
}
