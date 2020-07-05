package structs

import (
	"time"
	"unsafe"

	"github.com/erikdubbelboer/ip-api-proxy/field"
	_ "github.com/mailru/easyjson/gen"
)

//go:generate easyjson

//easyjson:json
type Response struct {
	Status        string  `json:"status,omitempty"`        // "success"
	Continent     string  `json:"continent,omitempty"`     // "North America"
	ContinentCode string  `json:"continentCode,omitempty"` // "NA"
	Country       string  `json:"country,omitempty"`       // "Canada"
	CountryCode   string  `json:"countryCode,omitempty"`   // "CA"
	Region        string  `json:"region,omitempty"`        // "QC"
	RegionName    string  `json:"regionName,omitempty"`    // "Quebec"
	City          string  `json:"city,omitempty"`          // "Montreal"
	District      string  `json:"district,omitempty"`      // """"
	Zip           string  `json:"zip,omitempty"`           // "H1S"
	Lat           float64 `json:"lat,omitempty"`           // 45.5808
	Lon           float64 `json:"lon,omitempty"`           // -73.5825
	Timezone      string  `json:"timezone,omitempty"`      // "America/Toronto"
	Offset        int     `json:"offset,omitempty"`        // -14400
	Currency      string  `json:"currency,omitempty"`      // "CAD"
	ISP           string  `json:"isp,omitempty"`           // "Le Groupe Videotron Ltee"
	Org           string  `json:"org,omitempty"`           // "Videotron Ltee"
	AS            string  `json:"as,omitempty"`            // "AS5769 Videotron Telecom Ltee"
	ASName        string  `json:"asname,omitempty"`        // "VIDEOTRON"
	Reverse       string  `json:"reverse,omitempty"`       // "modemcable001.0-48-24.mc.videotron.ca"
	Mobile        *bool   `json:"mobile,omitempty"`        // false
	Proxy         *bool   `json:"proxy,omitempty"`         // false
	Hosting       *bool   `json:"hosting,omitempty"`       // false
	Message       string  `json:"message,omitempty"`       // "invalid query"
	Query         string  `json:"query,omitempty"`         // "24.48.0.1"
}

func (r Response) Trim(fields field.Fields) Response {
	if !fields.Contains(16384) {
		r.Status = ""
	}
	if !fields.Contains(1048576) {
		r.Continent = ""
	}
	if !fields.Contains(2097152) {
		r.ContinentCode = ""
	}
	if !fields.Contains(1) {
		r.Country = ""
	}
	if !fields.Contains(2) {
		r.CountryCode = ""
	}
	if !fields.Contains(4) {
		r.Region = ""
	}
	if !fields.Contains(8) {
		r.RegionName = ""
	}
	if !fields.Contains(16) {
		r.City = ""
	}
	if !fields.Contains(524288) {
		r.District = ""
	}
	if !fields.Contains(32) {
		r.Zip = ""
	}
	if !fields.Contains(64) {
		r.Lat = 0
	}
	if !fields.Contains(128) {
		r.Lon = 0
	}
	if !fields.Contains(256) {
		r.Timezone = ""
	}
	if !fields.Contains(33554432) {
		r.Offset = 0
	}
	if !fields.Contains(8388608) {
		r.Currency = ""
	}
	if !fields.Contains(512) {
		r.ISP = ""
	}
	if !fields.Contains(1024) {
		r.Org = ""
	}
	if !fields.Contains(2048) {
		r.AS = ""
	}
	if !fields.Contains(4194304) {
		r.ASName = ""
	}
	if !fields.Contains(4096) {
		r.Reverse = ""
	}
	if !fields.Contains(65536) {
		r.Mobile = nil
	}
	if !fields.Contains(131072) {
		r.Proxy = nil
	}
	if !fields.Contains(16777216) {
		r.Hosting = nil
	}
	if !fields.Contains(32768) {
		r.Message = ""
	}
	if !fields.Contains(8192) {
		r.Query = ""
	}
	return r
}

//easyjson:json
type Responses []Response

//easyjson:json
type CacheEntry struct {
	IP       string       `json:"query"`
	Lang     string       `json:"lang"`
	Fields   field.Fields `json:"fields"`
	Expires  time.Time    `json:"-"`
	Response Response     `json:"-"`
}

var emptyCacheEntrySize = int(unsafe.Sizeof(CacheEntry{}))

// Size returns the size of the CacheEntry in bytes.
func (c *CacheEntry) Size() int {
	size := emptyCacheEntrySize +
		len(c.IP) +
		len(c.Lang) +
		len(c.Response.Status) +
		len(c.Response.Continent) +
		len(c.Response.ContinentCode) +
		len(c.Response.Country) +
		len(c.Response.CountryCode) +
		len(c.Response.Region) +
		len(c.Response.RegionName) +
		len(c.Response.City) +
		len(c.Response.District) +
		len(c.Response.Zip) +
		len(c.Response.Timezone) +
		len(c.Response.Currency) +
		len(c.Response.ISP) +
		len(c.Response.Org) +
		len(c.Response.AS) +
		len(c.Response.ASName) +
		len(c.Response.Reverse) +
		len(c.Response.Message) +
		len(c.Response.Query)

	if c.Response.Mobile != nil {
		size += 1
	}
	if c.Response.Proxy != nil {
		size += 1
	}
	if c.Response.Hosting != nil {
		size += 1
	}

	return size
}

//easyjson:json
type CacheEntries []*CacheEntry
