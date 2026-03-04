package weathergov_test

import (
	"bytes"
	"encoding/xml"
	"os"
	"testing"
	"time"

	"github.com/chaimleib/weatherfmt/weathergov"
	"github.com/stretchr/testify/assert"
)

var PortlandData = func() []byte {
	f, err := os.ReadFile("testdata/portland.xml")
	if err != nil {
		panic(err)
	}
	return f
}()

func TestDWML(t *testing.T) {
	var got weathergov.DWML
	dec := xml.NewDecoder(bytes.NewReader(PortlandData))
	dec.CharsetReader = weathergov.CharsetReader
	err := dec.Decode(&got)
	assert.NoError(t, err)

	// dwml
	assert.Equal(t, "1.0", got.Version)
	assert.Equal(t, "http://www.w3.org/2001/XMLSchema", got.XSD)
	assert.Equal(t, "http://www.w3.org/2001/XMLSchema-instance", got.XSI)
	assert.Equal(
		t,
		"http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd",
		got.NoNamespaceSchemaLocation,
	)

	// dwml>head>product
	assert.Equal(t, "tabular-digital", got.Head.Product.ConciseName)
	assert.Equal(t, "developmental", got.Head.Product.OperationalMode)
	assert.Equal(t, "WGS 1984", got.Head.Product.SRSName)

	// dwml>head>product>creation-date
	wantTZ := time.FixedZone("", -8*int(time.Hour/time.Second))
	wantCreationDate := time.Date(
		2026, time.March, 1,
		1, 1, 52, 0,
		wantTZ,
	)
	assert.Equal(t, wantCreationDate, got.Head.Product.Creation.Date)
	assert.Equal(t, "PT1H", got.Head.Product.Creation.RefreshFrequency)

	// dwml>head>source
	assert.Equal(t, "Portland, OR", got.Head.Source.ProductionCenter)
	assert.Equal(t, "https://www.weather.gov/pqr", got.Head.Source.Credit)
	assert.Equal(
		t,
		"https://www.nws.noaa.gov/forecasts/xml/",
		got.Head.Source.MoreInformation,
	)

	// dwml>data>location
	assert.Equal(t, "point1", got.Data.Location.LocationKey)
	assert.Equal(t, "Portland OR, OR", got.Data.Location.Description)

	// dwml>data>location>point
	assert.InDelta(t, 45.53, got.Data.Location.Point.Latitude, 0.001)
	assert.InDelta(t, -122.67, got.Data.Location.Point.Longitude, 0.001)
	
	// dwml>data>location>city
	assert.Equal(t, "OR", got.Data.Location.City.State)
	assert.Equal(t, "Portland OR", got.Data.Location.City.Name)

	// dwml>data>location>height
	assert.Equal(t, "mean sea level", got.Data.Location.Height.Datum)
	assert.Equal(t, 200, got.Data.Location.Height.Value)

	// dwml>data>moreWeatherInformation
	assert.Equal(
		t,
		"point1",
		got.Data.MoreWeatherInformation.ApplicableLocation,
	)
	assert.Equal(
		t,
		"//forecast.weather.gov/MapClick.php?lat=45.53&lon=-122.67&FcstType=digital",
		got.Data.MoreWeatherInformation.Value,
	)

	// dwml>data>time-layout
	assert.Equal(t, "local", got.Data.TimeLayout.TimeCoordinate)
	assert.Equal(t, "none", got.Data.TimeLayout.Summarization)
	assert.Equal(t, "k-p1h-n1-0", got.Data.TimeLayout.LayoutKey)
	wantStarts := make([]time.Time, 0, 7*24)
	wantEnds := make([]time.Time, 0, cap(wantStarts))
	thisStart := time.Date(
		2026, time.March, 1,
		1, 0, 0, 0,
		wantTZ,
	)
	for range cap(wantStarts) {
		wantStarts = append(wantStarts, thisStart)
		thisStart = time.Date(
			thisStart.Year(), thisStart.Month(), thisStart.Day(),
			thisStart.Hour()+1, 0, 0, 0,
			wantTZ,
		)
		wantEnds = append(wantEnds, thisStart)
	}
	assert.Equal(t, wantStarts, got.Data.TimeLayout.StartValidTime)
	assert.Equal(t, wantEnds, got.Data.TimeLayout.EndValidTime)

	// dwml>data>parameters
	assert.Equal(t, "point1", got.Data.Parameters.ApplicableLocation)

	// dwml>data>parameters>*
	wantTimeSeries := []struct{
		Name string
		Type string
		Units string
		Values []string
	}{
		{
			Name: "temperature",
			Type: "hourly",
			Values: []string{
				"44", "43", "43", "43", "42", "41", 
				"42", "43", "46", "49", "52", "56",

				"57", "58", "59", "58", "57", "54", 
				"52", "51", "50", "49", "49", "48",

				"47", "46", "46", "45", "45", "44", 
				"44", "46", "48", "50", "53", "55",

				"58", "59", "59", "58", "57", "55", 
				"53", "52", "51", "50", "49", "49",

				"48", "47", "47", "47", "46", "46", 
				"46", "47", "49", "50", "51", "53",

				"53", "54", "54", "53", "53", "52", 
				"51", "50", "50", "50", "50", "50",

				"50", "50", "49", "49", "48", "48", 
				"47", "48", "49", "49", "51", "52",

				"52", "52", "52", "50", "49", "48", 
				"47", "47", "46", "46", "45", "45",

				"44", "44", "44", "44", "43", "43", 
				"43", "44", "45", "47", "49", "50",

				"51", "51", "51", "50", "49", "49", 
				"48", "47", "47", "46", "46", "46",

				"46", "46", "45", "45", "45", "45", 
				"45", "46", "47", "49", "51", "52",

				"53", "53", "53", "52", "51", "50", 
				"49", "48", "48", "47", "47", "46",

				"46", "46", "45", "45", "45", "45", 
				"45", "46", "47", "48", "50", "52",

				"53", "53", "53", "52", "51", "51", 
				"50", "50", "49", "49", "49", "48",
			},
		},
		{
			Name: "temperature",
			Type: "dew point",
			Values: []string{
				"40", "39", "39", "39", "38", "37", 
				"38", "39", "40", "41", "43", "45",
				
				"44", "45", "46", "45", "47", "46", 
				"45", "45", "45", "45", "46", "46",
				
				"45", "44", "44", "44", "44", "43", 
				"42", "44", "46", "47", "49", "49",
				
				"51", "51", "50", "50", "50", "50", 
				"50", "49", "48", "48", "47", "47",
				
				"46", "45", "45", "45", "44", "45", 
				"45", "46", "47", "48", "48", "50",
				
				"49", "50", "49", "48", "49", "48", 
				"48", "47", "48", "48", "48", "48",
				
				"48", "48", "47", "47", "46", "46", 
				"45", "46", "47", "47", "48", "48",
				
				"47", "47", "47", "45", "45", "44", 
				"44", "45", "44", "44", "43", "44",
				
				"43", "43", "42", "42", "41", "42", 
				"42", "43", "43", "44", "45", "45",
				
				"46", "46", "46", "45", "44", "44", 
				"44", "44", "45", "44", "44", "44",
				
				"44", "44", "43", "43", "43", "44", 
				"44", "45", "45", "46", "46", "45",
				
				"45", "45", "45", "45", "44", "43", 
				"43", "43", "43", "43", "43", "43",
				
				"43", "44", "43", "43", "43", "43", 
				"43", "44", "45", "45", "46", "47",
				
				"47", "47", "47", "47", "47", "48", 
				"47", "47", "46", "46", "46", "46",
			},
		},
		{
			Name: "temperature",
			Type: "wind chill",
			Values: []string{
				"44", "43", "43", "43", "42", "41",
				"42", "43", "46", "48", "", "", // 11
				"", "", "", "", "", "",
				"", "", "50", "49", "49", "48", // 23
				"47", "46", "46", "45", "45", "44",
				"44", "46", "48", "50", "", "", // 35
				"", "", "", "", "", "",
				"", "", "", "50", "49", "49", // 47
				"48", "47", "47", "46", "45", "45",
				"44", "45", "47", "47", "", "", // 59
				"", "", "", "", "", "",
				"", "47", "47", "47", "47", "47", // 71
				"46", "46", "45", "46", "44", "44",
				"43", "44", "46", "44", "", "", // 83
				"", "", "", "46", "45", "44",
				"44", "44", "42", "43", "42", "42", // 95
				"41", "41", "41", "41", "39", "39",
				"39", "41", "42", "43", "46", "47", // 107
				"", "", "", "47", "46", "46",
				"45", "44", "44", "43", "43", "43", // 119
				"43", "43", "42", "43", "43", "43",
				"42", "43", "44", "46", "", "", // 131
				"", "", "", "", "", "47",
				"47", "45", "45", "44", "44", "43", // 143
				"43", "43", "42", "42", "42", "42",
				"42", "43", "44", "44", "47", "", // 155
				"", "", "", "", "", "",
				"47", "47", "46", "47", "47", "45", // 167
			},
		},
		{
			Name: "probability-of-precipitation",
			Type: "floating",
			Units: "percent",
			Values: []string {
				"5", "7", "7", "5", "5", "6", 
				"6", "7", "8", "12", "11", "11", 
				
				"11", "12", "13", "16", "17", "19", 
				"19", "19", "21", "12", "12", "13", 
				
				"11", "10", "10", "14", "14", "14", 
				"14", "14", "14", "15", "15", "15", 
				
				"15", "15", "15", "18", "18", "18", 
				"18", "18", "18", "9", "9", "9", 
				
				"9", "9", "9", "25", "25", "25", 
				"25", "25", "25", "71", "71", "71", 
				
				"71", "71", "71", "88", "88", "88", 
				"88", "88", "88", "100", "100", "100", 
				
				"100", "100", "100", "100", "100", "100", 
				"100", "100", "100", "99", "99", "99", 
				
				"99", "99", "99", "94", "94", "94", 
				"94", "94", "94", "67", "67", "67", 
				
				"67", "67", "67", "79", "79", "79", 
				"79", "79", "79", "83", "83", "83", 
				
				"83", "83", "83", "72", "72", "72", 
				"72", "72", "72", "53", "53", "53", 
				
				"53", "53", "53", "56", "56", "56", 
				"56", "56", "56", "57", "57", "57", 
				
				"57", "57", "57", "61", "61", "61", 
				"61", "61", "61", "53", "53", "53", 
				
				"53", "53", "53", "57", "57", "57", 
				"57", "57", "57", "64", "64", "64", 
				
				"64", "64", "64", "71", "71", "71", 
				"71", "71", "71", "58", "58", "58", 
			},
		},
		{
			Name: "wind-speed",
			Type: "sustained",
			Values: []string {
				"2", "2", "2", "2", "2", "2", 
				"2", "2", "2", "3", "3", "3", 
				
				"3", "3", "3", "3", "2", "2", 
				"2", "2", "2", "2", "2", "2", 
				
				"2", "2", "2", "1", "1", "1", 
				"2", "2", "2", "2", "2", "2", 
				
				"5", "5", "5", "5", "5", "5", 
				"2", "2", "2", "2", "2", "2", 
				
				"2", "2", "2", "3", "3", "3", 
				"5", "5", "5", "7", "7", "7", 
				
				"10", "10", "10", "9", "9", "9", 
				"8", "8", "8", "8", "8", "8", 
				
				"9", "9", "9", "8", "8", "8", 
				"8", "8", "8", "11", "11", "11", 
				
				"11", "11", "11", "9", "9", "9", 
				"7", "7", "7", "6", "6", "6", 
				
				"6", "6", "6", "6", "6", "6", 
				"6", "6", "6", "8", "8", "8", 
				
				"9", "9", "9", "8", "8", "8", 
				"6", "6", "6", "6", "6", "6", 
				
				"6", "6", "6", "5", "5", "5", 
				"6", "6", "6", "7", "7", "7", 
				
				"9", "9", "9", "7", "7", "7", 
				"6", "6", "6", "6", "6", "6", 
				
				"6", "6", "6", "6", "6", "6", 
				"6", "6", "6", "8", "8", "8", 
				
				"9", "9", "9", "8", "8", "8", 
				"7", "7", "7", "6", "6", "6", 
			},
		},
		{
			Name: "wind-speed",
			Type: "gust",
			Values: []string{
				"", "", "", "", "", "", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "",  

				"20", "20", "20", "18", "18", "18", 
				"20", "20", "20", "21", "21", "21", 

				"21", "21", "21", "20", "20", "20", 
				"18", "18", "18", "22", "22", "22", 

				"23", "23", "23", "20", "20", "20", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "",  

				"18", "18", "18", "", "", "", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "", 

				"", "", "", "", "", "", 
				"", "", "", "", "", "",  

				"20", "20", "20", "", "", "", 
				"", "", "", "", "", "", 
			},
		},
		{
			Name: "direction",
			Type: "wind",
			Units: "degrees true",
			Values: []string{
				"350", "0", "0", "0", "0", "0", 
				"350", "350", "10", "50", "60", "70", 

				"80", "80", "80", "80", "120", "120", 
				"0", "50", "50", "0", "90", "140", 

				"270", "270", "320", "340", "340", "340", 
				"0", "0", "0", "180", "180", "180", 

				"210", "210", "210", "180", "180", "180", 
				"170", "170", "170", "150", "150", "150", 

				"170", "170", "170", "160", "160", "160", 
				"180", "180", "180", "190", "190", "190", 

				"200", "200", "200", "200", "200", "200", 
				"190", "190", "190", "180", "180", "180", 

				"190", "190", "190", "200", "200", "200", 
				"200", "200", "200", "210", "210", "210", 

				"230", "230", "230", "240", "240", "240", 
				"230", "230", "230", "220", "220", "220", 

				"210", "210", "210", "220", "220", "220", 
				"200", "200", "200", "210", "210", "210", 

				"220", "220", "220", "220", "220", "220", 
				"210", "210", "210", "200", "200", "200", 

				"210", "210", "210", "210", "210", "210", 
				"210", "210", "210", "230", "230", "230", 

				"240", "240", "240", "250", "250", "250", 
				"230", "230", "230", "210", "210", "210", 

				"200", "200", "200", "190", "190", "190", 
				"190", "190", "190", "190", "190", "190", 

				"190", "190", "190", "200", "200", "200", 
				"200", "200", "200", "210", "210", "210", 
			},
		},
		{
			Name: "cloud-amount",
			Type: "total",
			Units: "percent",
			Values: []string{
				"80", "75", "79", "76", "78", "69", 
				"74", "70", "72", "71", "67", "70", 

				"71", "72", "69", "79", "82", "84", 
				"74", "83", "81", "73", "84", "81", 

				"77", "84", "70", "60", "60", "60", 
				"71", "71", "71", "75", "75", "75", 

				"55", "55", "55", "33", "33", "33", 
				"33", "33", "33", "70", "70", "70", 

				"87", "87", "87", "87", "87", "87", 
				"89", "89", "89", "88", "88", "88", 

				"87", "87", "87", "93", "93", "93", 
				"92", "92", "92", "100", "100", "100", 

				"100", "100", "100", "100", "100", "100", 
				"100", "100", "100", "100", "100", "100", 

				"100", "100", "100", "96", "96", "96", 
				"96", "96", "96", "78", "78", "78", 

				"87", "87", "87", "86", "86", "86", 
				"91", "91", "91", "89", "89", "89", 

				"89", "89", "89", "90", "90", "90", 
				"89", "89", "89", "83", "83", "83", 

				"87", "87", "87", "87", "87", "87", 
				"91", "91", "91", "89", "89", "89", 

				"87", "87", "87", "87", "87", "87", 
				"87", "87", "87", "82", "82", "82", 

				"83", "83", "83", "83", "83", "83", 
				"90", "90", "90", "89", "89", "89", 

				"80", "80", "80", "81", "81", "81", 
				"81", "81", "81", "82", "82", "82", 
			},
		},
		{
			Name: "humidity",
			Type: "relative",
			Units: "percent",
			Values: []string{
				"86", "85", "87", "87", "87", "87", 
				"87", "86", "80", "75", "70", "66", 

				"61", "61", "62", "62", "68", "74", 
				"78", "81", "83", "85", "90", "91", 

				"91", "94", "93", "96", "95", "95", 
				"94", "93", "91", "89", "85", "81", 

				"77", "74", "73", "74", "78", "83", 
				"88", "90", "91", "92", "93", "94", 

				"94", "94", "94", "94", "94", "95", 
				"95", "95", "94", "93", "91", "89", 

				"87", "85", "84", "84", "85", "86", 
				"88", "90", "92", "93", "94", "94", 

				"94", "94", "94", "94", "94", "93", 
				"93", "94", "94", "94", "91", "87", 

				"84", "83", "84", "83", "85", "87", 
				"89", "91", "92", "93", "94", "96", 

				"96", "95", "94", "93", "94", "95", 
				"96", "95", "92", "89", "87", "84", 

				"83", "82", "82", "83", "83", "84", 
				"86", "88", "91", "93", "93", "93", 

				"93", "92", "92", "93", "94", "96", 
				"96", "95", "93", "89", "84", "78", 

				"74", "74", "75", "77", "78", "78", 
				"80", "82", "84", "86", "87", "88", 

				"89", "91", "92", "93", "93", "93", 
				"93", "92", "91", "89", "86", "82", 

				"80", "80", "81", "83", "85", "88", 
				"89", "90", "89", "89", "90", "91", 
			},
		},
		{
			Name: "hourly-qpf",
			Type: "floating",
			Units: "inches",
			Values: []string{
				"0", "0", "0", "0", "0", "0", 
				"0", "0", "0", "0", "0", "0", 

				"0", "0", "0", "0", "0", "0", 
				"0", "0", "0", "0", "0", "0", 

				"0", "0", "0", "0", "0", "0", 
				"0", "0", "0", "0", "0", "0", 

				"0", "0", "0", "0", "0", "0", 
				"0", "0", "0", "0", "0", "0", 

				"0", "0", "0", "0", "0", "0", 
				"0", "0", "0", "0.0067", "0.0067", "0.0067", 

				"0.0067", "0.0067", "0.0067", "0.0167", "0.0167", "0.0167", 
				"0.0167", "0.0167", "0.0167", "0.0317", "0.0317", "0.0317", 

				"0.0317", "0.0317", "0.0317", "0.0367", "0.0367", "0.0367", 
				"0.0367", "0.0367", "0.0367", "0.0300", "0.0300", "0.0300", 

				"0.0300", "0.0300", "0.0300", "0.0100", "0.0100", "0.0100", 
				"0.0100", "0.0100", "0.0100", "0.0050", "0.0050", "0.0050", 

				"0.0050", "0.0050", "0.0050", "0.0050", "0.0050", "0.0050", 
				"0.0050", "0.0050", "0.0050", "0.0100", "0.0100", "0.0100", 

				"0.0100", "0.0100", "0.0100", "0.0083", "0.0083", "0.0083", 
				"0.0083", "0.0083", "0.0083", "0.0050", "0.0050", "0.0050", 

				"0.0050", "0.0050", "0.0050", "0.0033", "0.0033", "0.0033", 
				"0.0033", "0.0033", "0.0033", "0.0050", "0.0050", "0.0050", 

				"0.0050", "0.0050", "0.0050", "0.0033", "0.0033", "0.0033", 
				"0.0033", "0.0033", "0.0033", "0.0033", "0.0033", "0.0033", 

				"0.0033", "0.0033", "0.0033", "0.0067", "0.0067", "0.0067", 
				"0.0067", "0.0067", "0.0067", "0.0100", "0.0100", "0.0100", 

				"0.0100", "0.0100", "0.0100", "0.0100", "0.0100", "0.0100", 
				"0.0100", "0.0100", "0.0100", "0.0067", "0.0067", "0.0067", 
			},
		},
	}
	assert.Len(t, got.Data.Parameters.TimeSeries, 10)
	for i, gotTS := range got.Data.Parameters.TimeSeries {
		if i >= len(wantTimeSeries) {
			t.Errorf(
				"got more TimeSeries (%d) than wanted (%d)",
				len(got.Data.Parameters.TimeSeries),
				len(wantTimeSeries),
			)
			break
		}
		want := wantTimeSeries[i]
		assert.Equal(t, want.Name, gotTS.Name())
		assert.Equal(t, want.Type, gotTS.Type)
		assert.Equal(t, "k-p1h-n1-0", gotTS.TimeLayout)
		assert.Equal(t, want.Values, gotTS.Values.Strings(), want.Type)
	}

	assert.Equal(t, "k-p1h-n1-0", got.Data.Parameters.Weather.TimeLayout)
	wantConditions := []weathergov.WeatherConditionsValues{
		nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, // 11

		nil, nil, nil,
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		nil, nil, nil, // 23
		
		nil, nil, nil, nil, nil, nil,
		nil, nil, nil,
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}}, // 35

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "slight chance"}},
		nil, nil, nil, // 47

		nil, nil, nil,
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}}, // 59

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}}, // 71


		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}}, // 83

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}}, // 95

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}}, // 107

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: ""}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}}, // 119

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}}, // 131

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}}, // 143

		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "chance"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}}, // 155

		
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}},
		[]weathergov.WeatherConditionsValue{{WeatherType: "rain", Coverage: "likely"}}, // 167
	}
	gotConditions := got.Data.Parameters.Weather.WeatherConditions
	assert.Len(t, gotConditions, len(wantConditions))
	for i, want := range wantConditions {
		if i >= len(gotConditions) {
			break
		}
		if want == nil {
			assert.True(t, gotConditions[i].Nil, "weather-condiitions[%d].Nil", i)
		} else {
			assert.False(t, gotConditions[i].Nil, "weather-condiitions[%d].Nil", i)
			assert.Equal(t, want, gotConditions[i].Values,
				"weather-conditions[%d].Values", i)
		}
	}
}
