package satellite

type GPElement struct {
	ObjectName      string  `json:"OBJECT_NAME"`
	ObjectID        string  `json:"OBJECT_ID"`
	NoradCatID      int     `json:"NORAD_CAT_ID"`
	Epoch           string  `json:"EPOCH"`
	MeanMotion      float64 `json:"MEAN_MOTION"`
	Eccentricity    float64 `json:"ECCENTRICITY"`
	Inclination     float64 `json:"INCLINATION"`
	RAOfAscNode     float64 `json:"RA_OF_ASC_NODE"`
	ArgOfPericenter float64 `json:"ARG_OF_PERICENTER"`
	MeanAnomaly     float64 `json:"MEAN_ANOMALY"`
}

type Observer struct {
	Name            string
	LatitudeDeg     float64
	LongitudeDeg    float64
	ElevationMeters float64
}

type ApparentPosition struct {
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	Source        string `json:"source"`
	NoradID       int    `json:"norad_id"`
	ObjectID      string `json:"object_id"`
	ObserverName  string `json:"observer_name"`
	TimeUTC       string `json:"time_utc"`

	AzimuthDeg   float64 `json:"azimuth_deg"`
	ElevationDeg float64 `json:"elevation_deg"`
	RangeKm      float64 `json:"range_km"`
	RangeRateKms float64 `json:"range_rate_km_s"`
	Visible      bool    `json:"visible"`

	SatelliteLatitudeDeg  float64 `json:"satellite_latitude_deg"`
	SatelliteLongitudeDeg float64 `json:"satellite_longitude_deg"`
	SatelliteAltitudeKm   float64 `json:"satellite_altitude_km"`
}
