package weathergov

import (
	"time"

	"github.com/chaimleib/weatherfmt/calendar"
)

const URLTemplate = "https://forecast.weather.gov/MapClick.php?lat=%.2f&lon=%.2f&FcstType=digitalDWML"

const Period = time.Hour

var TimeSpan = calendar.Duration{Days: 14}
