package shared

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoLocation struct {
	City          string
	Country       string
	CountryCode   string
	Continent     string
	ContinentCode string
}

func LookupIP(mmdbPath, ipAddress string) (*GeoLocation, error) {
	db, err := geoip2.Open(mmdbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open mmdb: %w", err)
	}
	defer db.Close()

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipAddress)
	}

	record, err := db.City(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup IP: %w", err)
	}

	country := record.Country.Names["en"]
	countryCode := record.Country.IsoCode

	if country == "" && countryCode == "" {
		return nil, fmt.Errorf("no location data found for IP: %s", ipAddress)
	}

	return &GeoLocation{
		City:          record.City.Names["en"],
		Country:       country,
		CountryCode:   countryCode,
		Continent:     record.Continent.Names["en"],
		ContinentCode: record.Continent.Code,
	}, nil
}
