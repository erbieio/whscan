package utils

import (
	_ "embed"
	"net"

	"github.com/oschwald/geoip2-golang"
)

//go:embed GeoLite2-City.mmdb
var geoLite2 []byte
var geoDB, _ = geoip2.FromBytes(geoLite2)

func IP2Location(ip string) (name string, latitude, longitude float64) {
	city, err := geoDB.City(net.ParseIP(ip))
	if err != nil {
		return
	}
	language := "en"
	name += city.Country.Names[language] + ","
	if len(city.Subdivisions) > 0 {
		name += city.Subdivisions[0].Names[language] + ","
	} else {
		name += city.City.Names[language] + ","
	}
	name += city.City.Names[language]
	latitude = city.Location.Latitude
	longitude = city.Location.Longitude
	return
}
