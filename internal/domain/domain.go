package domain

import (
	"encoding/json"
	"net"
	"time"
)

type Event struct {
	// If the JSON looks like {"ip_address": "192.168.1.1"}
	IPAddress net.IP `json:"ip_address"`
}

type Location struct {
	IPAddress net.IP           `json:"ip"`
	Country   string           `json:"country"`
	City      string           `json:"city"`
	Continent string           `json:"continent"`
	Location  LocationMetadata `json:"location"`
}

type LocationMetadata struct {
	Latitude       float32  `json:"latitude"`
	Longitude      float32  `json:"longitude"`
	AccuracyRadius int      `json:"accuracy_radius"`
	TimeZone       TimeZone `json:"time_zone"`
}

type TimeZone struct {
	*time.Location
}

func (tz *TimeZone) UnmarshalJSON(b []byte) error {
	var zoneName string
	if err := json.Unmarshal(b, &zoneName); err != nil {
		return err
	}

	loc, err := time.LoadLocation(zoneName)
	if err != nil {
		return err
	}
	tz.Location = loc
	return nil
}
