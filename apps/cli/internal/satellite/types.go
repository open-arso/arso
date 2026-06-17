package satellite

import (
    "time"
)

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

	EphemerisType      int     `json:"EPHEMERIS_TYPE,omitempty"`
	ClassificationType string `json:"CLASSIFICATION_TYPE,omitempty"`
	ElementSetNo       int     `json:"ELEMENT_SET_NO,omitempty"`
	RevAtEpoch         int     `json:"REV_AT_EPOCH,omitempty"`
	BStar              float64 `json:"BSTAR,omitempty"`
	MeanMotionDot      float64 `json:"MEAN_MOTION_DOT,omitempty"`
	MeanMotionDDot     float64 `json:"MEAN_MOTION_DDOT,omitempty"`
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
	AboveHorizon bool    `json:"above_horizon"`

	SatelliteLatitudeDeg  float64 `json:"satellite_latitude_deg"`
	SatelliteLongitudeDeg float64 `json:"satellite_longitude_deg"`
	SatelliteAltitudeKm   float64 `json:"satellite_altitude_km"`
}

type ResolvedTarget struct {
	Query      string    `json:"query"`
	Name       string    `json:"name"`
	ObjectID   string    `json:"objectId"`
	NoradID    int       `json:"noradId"`
	Kind       string    `json:"kind"`
	Source     string    `json:"source"`
	ResolvedAt time.Time `json:"resolvedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type PassPredictionResult struct {
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	Source        string `json:"source"`
	NoradID       int    `json:"norad_id"`
	ObjectID      string `json:"object_id"`
	ObserverName  string `json:"observer_name"`

	Passes []PredictedPass `json:"passes"`
}

type PredictedPass struct {
	AcquisitionOfSignal   time.Time  	 `json:"acquisition_of_signal"`
	LossOfSignal 		  time.Time  	 `json:"loss_of_signal"`
	Duration 			  time.Duration  `json:"duration"`
	MaxElevation 		  float64  		 `json:"max_elevation"`
	MaxElevationTime      time.Time 	 `json:"time_of_max_elevation"`   
	AzimuthAtAOS		  float64  		 `json:"azimuth_at_aos"`
	AzimuthAtLOS		  float64  		 `json:"azimuth_at_los"`
}

