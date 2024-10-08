package activity

type Header struct {
	Meta     HeaderMeta     `yaml:"meta"`
	Activity HeaderActivity `yaml:"activity"`
	Layout   Layout         `yaml:"layout"`
	Stats    Stats          `yaml:"stats"`
}

type HeaderMeta struct {
	Version string `yaml:"version,omitempty"`
}

type HeaderActivity struct {
	Wandern       bool          `yaml:"wandern,omitempty"`
	Type          string        `yaml:"type"`
	Date          string        `yaml:"date"`
	Title         string        `yaml:"title"`
	PointOfOrigin PointOfOrigin `yaml:"pointOfOrigin"`
	Season        string        `yaml:"season"`
	Rating        string        `yaml:"rating"`
	Company       string        `yaml:"company"`
	Restaurant    string        `yaml:"restaurant"`
	MaxElevation  string        `yaml:"maxElevation"`
}

type PointOfOrigin struct {
	Name   string `yaml:"name"`
	Qr     string `yaml:"qr"`
	Region string `yaml:"region"`
}

type Stats struct {
	Ascent      string `yaml:"ascent"`
	Distance    string `yaml:"distance"`
	MovingTime  string `yaml:"movingTime"`
	OverallTime string `yaml:"overallTime"`
	StartTime   string `yaml:"startTime"`
	SummitTime  string `yaml:"summitTime"`
	Puls        string `yaml:"puls,omitempty"`
}

type Layout struct {
	HeadElevationProfile        bool    `yaml:"headElevationProfile"`
	ElevationProfileType        string  `yaml:"elevationProfileType,omitempty"`
	ElevationProfileRightMargin float32 `yaml:"elevationProfileRightMargin"`
	TableSize                   float32 `yaml:"tableSize"`
	MapSize                     float32 `yaml:"mapSize"`
	MapHeight                   int     `yaml:"mapHeight"`
	Linespread                  float32 `yaml:"linespread"`
}
