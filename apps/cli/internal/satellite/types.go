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

