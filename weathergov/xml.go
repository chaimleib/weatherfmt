package weathergov

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"

	"golang.org/x/text/encoding/charmap"
)

// XML structure for data downloaded from forecast.weather.gov/MapClick.php.

type DWML struct {
	XMLName                   struct{} `xml:"dwml"`
	Version                   string   `xml:"version,attr"`
	XSD                       string   `xml:"xsd,attr"`
	XSI                       string   `xml:"xsi,attr"`
	NoNamespaceSchemaLocation string   `xml:"noNamespaceSchemaLocation,attr"`
	Head                      Head
	Data                      Data
}

type Head struct {
	XMLName struct{} `xml:"head"`
	Product Product
	Source  Source
}

type Product struct {
	XMLName         struct{} `xml:"product"`
	ConciseName     string   `xml:"concise-name,attr"     example:"tabular-digital"`
	OperationalMode string   `xml:"operational-mode,attr" example:"developmental"`
	SRSName         string   `xml:"srsName,attr"          example:"WGS 1984"`
	Creation        Creation
}

type Creation struct {
	XMLName          struct{}  `xml:"creation-date"`
	Date             time.Time `xml:",chardata"              example:"2026-03-01T01:01:52-08:00"`
	RefreshFrequency string    `xml:"refresh-frequency,attr" example:"PT1H"`
}

type Source struct {
	XMLName          struct{} `xml:"source"`
	ProductionCenter string   `xml:"production-center" example:"Portland, OR"`
	Credit           string   `xml:"credit"            example:"https://www.weater.gov/pqr"`
	MoreInformation  string   `xml:"more-information"  example:"https://www.nws.noaa.gov/forecasts/xml/"`
}

// Data is located at dwml>data
type Data struct {
	XMLName                struct{} `xml:"data"`
	Location               Location
	MoreWeatherInformation MoreWeatherInformation
	TimeLayout             TimeLayout
	Parameters             Parameters
}

type Location struct {
	XMLName     struct{} `xml:"location"`
	LocationKey string   `xml:"location-key" example:"point1"`
	Description string   `xml:"description"  example:"Portland OR, OR"`
	Point       Point
	City        City
	Height      Height
}

type Point struct {
	XMLName   struct{} `xml:"point"`
	Latitude  float64  `xml:"latitude,attr"  example:"45.53"`
	Longitude float64  `xml:"longitude,attr" example:"-122.67"`
}

type City struct {
	XMLName struct{} `xml:"city"`
	State   string   `xml:"state,attr" example:"OR"`
	Name    string   `xml:",chardata"  example:"Portland OR"`
}

type Height struct {
	XMLName struct{} `xml:"height"`
	Datum   string   `xml:"datum,attr" example:"mean sea level"`
	Value   int      `xml:",chardata"`
}

// MoreWeatherInformation is located at dwml>data>moreWeatherInformation
type MoreWeatherInformation struct {
	XMLName            struct{} `xml:"moreWeatherInformation"`
	ApplicableLocation string   `xml:"applicable-location,attr" example:"point1"`
	Value              string   `xml:",chardata"                example:"//forecast.weather.gov/MapClick.php?lat=45.53&lon=-122.67&FcstType=digital"`
}

type TimeLayout struct {
	XMLName        struct{}    `xml:"time-layout"`
	TimeCoordinate string      `xml:"time-coordinate,attr" example:"local"`
	Summarization  string      `xml:"summarization,attr"   example:"none"`
	LayoutKey      string      `xml:"layout-key"           example:"k-p1h-n1-0"`
	StartValidTime []time.Time `xml:"start-valid-time"     example:"2026-03-01T01:00:00-08:00"`
	EndValidTime   []time.Time `xml:"end-valid-time"       example:"2026-03-01T02:00:00-08:00"`
}

type Parameters struct {
	XMLName            struct{}     `xml:"parameters"`
	ApplicableLocation string       `xml:"applicable-location,attr" example:"point1"`
	Weather            Weather      `xml:"weather"`
	TimeSeries         []TimeSeries `xml:",any"`
}

type TimeSeries struct {
	XMLName    xml.Name
	Type       string `xml:"type,attr"            example:"hourly"`
	Units      string `xml:"units,attr,omitempty" example:"percent"`
	TimeLayout string `xml:"time-layout,attr"     example:"k-p1h-n1-0"`
	Values     Values `xml:"value"`
}

func (s TimeSeries) Name() string { return s.XMLName.Local }

type Value struct {
	XMLName struct{} `xml:"value"`
	Value   string   `xml:",chardata"          example:"42"`
	Nil     bool     `xml:"nil,attr,omitempty" example:"true"`
}

type Values []Value

func (vals Values) Strings() []string {
	result := make([]string, 0, len(vals))
	for _, val := range vals {
		if val.Nil {
			result = append(result, "")
		} else {
			result = append(result, val.Value)
		}
	}
	return result
}

func (vals Values) IntPointers() []*int {
	result := make([]*int, 0, len(vals))
	for _, val := range vals {
		if val.Nil {
			result = append(result, nil)
		} else if i, err := strconv.Atoi(val.Value); err != nil {
			result = append(result, &i)
		} else {
			result = append(result, nil)
		}
	}
	return result
}

func (vals Values) Floats() []float64 {
	result := make([]float64, 0, len(vals))
	for _, val := range vals {
		if val.Nil {
			result = append(result, 0)
		} else if f, err := strconv.ParseFloat(val.Value, 64); err != nil {
			result = append(result, f)
		} else {
			result = append(result, 0)
		}
	}
	return result
}

type Weather struct {
	XMLName           struct{}            `xml:"weather"`
	TimeLayout        string              `xml:"time-layout,attr"   example:"k-p1h-n1-0"`
	WeatherConditions []WeatherConditions `xml:"weather-conditions"`
}

type WeatherConditions struct {
	XMLName struct{}                `xml:"weather-conditions"`
	Nil     bool                    `xml:"nil,attr,omitempty" example:"true"`
	Values  WeatherConditionsValues `xml:"value"`
}

type WeatherConditionsValues []WeatherConditionsValue

type WeatherConditionsValue struct {
	XMLName     struct{} `xml:"value"`
	WeatherType string   `xml:"weather-type,attr" example:"rain"`
	Coverage    string   `xml:"coverage,attr"     example:"slight chance"`
}

func CharsetReader(name string, r io.Reader) (io.Reader, error) {
	if name == "ISO-8859-1" {
		return charmap.ISO8859_1.NewDecoder().Reader(r), nil
	}
	return nil, fmt.Errorf("unknown encoding %q", name)
}
