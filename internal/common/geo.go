package common

import "net/http"

type Geo struct {
	City       string
	Country    string
	Continent  string
	Longitude  string
	Latitude   string
	Region     string
	RegionCode string
	MetroCode  string
	PostalCode string
	Timezone   string
}

func GetGeoFromHeaders(request *http.Request) *Geo {
	return &Geo{
		City:       request.Header.Get("cf-ipcity"),
		Country:    request.Header.Get("cf-ipcountry"),
		Continent:  request.Header.Get("cf-ipcontinent"),
		Longitude:  request.Header.Get("cf-iplongitude"),
		Latitude:   request.Header.Get("cf-iplatitude"),
		Region:     request.Header.Get("cf-region"),
		RegionCode: request.Header.Get("cf-region-code"),
		MetroCode:  request.Header.Get("cf-metro-code"),
		PostalCode: request.Header.Get("cf-postal-code"),
		Timezone:   request.Header.Get("cf-timezone"),
	}
}
