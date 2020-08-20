package structs

import (
	"time"
	"unsafe"

	_ "github.com/mailru/easyjson/gen"

	"github.com/ip-api/proxy/internal/field"
)

//go:generate easyjson

//easyjson:json
type Response struct {
	Status        *string  `json:"status,omitempty"`        // "success"
	Continent     *string  `json:"continent,omitempty"`     // "North America"
	ContinentCode *string  `json:"continentCode,omitempty"` // "NA"
	Country       *string  `json:"country,omitempty"`       // "Canada"
	CountryCode   *string  `json:"countryCode,omitempty"`   // "CA"
	Region        *string  `json:"region,omitempty"`        // "QC"
	RegionName    *string  `json:"regionName,omitempty"`    // "Quebec"
	City          *string  `json:"city,omitempty"`          // "Montreal"
	District      *string  `json:"district,omitempty"`      // """"
	Zip           *string  `json:"zip,omitempty"`           // "H1S"
	Lat           *float64 `json:"lat,omitempty"`           // 45.5808
	Lon           *float64 `json:"lon,omitempty"`           // -73.5825
	Timezone      *string  `json:"timezone,omitempty"`      // "America/Toronto"
	Offset        *int     `json:"offset,omitempty"`        // -14400
	Currency      *string  `json:"currency,omitempty"`      // "CAD"
	ISP           *string  `json:"isp,omitempty"`           // "Le Groupe Videotron Ltee"
	Org           *string  `json:"org,omitempty"`           // "Videotron Ltee"
	AS            *string  `json:"as,omitempty"`            // "AS5769 Videotron Telecom Ltee"
	ASName        *string  `json:"asname,omitempty"`        // "VIDEOTRON"
	Reverse       *string  `json:"reverse,omitempty"`       // "modemcable001.0-48-24.mc.videotron.ca"
	Mobile        *bool    `json:"mobile,omitempty"`        // false
	Proxy         *bool    `json:"proxy,omitempty"`         // false
	Hosting       *bool    `json:"hosting,omitempty"`       // false
	Message       *string  `json:"message,omitempty"`       // "invalid query"
	Query         *string  `json:"query,omitempty"`         // "24.48.0.1"
}

func ErrorResponse(status, message string) Response {
	return Response{
		// By default the response contains an upstream error.
		Status:  &status,
		Message: &message,
	}
}

func (r Response) Trim(fields field.Fields) Response {
	if !fields.Contains(16384) {
		r.Status = nil
	}
	if !fields.Contains(1048576) {
		r.Continent = nil
	}
	if !fields.Contains(2097152) {
		r.ContinentCode = nil
	}
	if !fields.Contains(1) {
		r.Country = nil
	}
	if !fields.Contains(2) {
		r.CountryCode = nil
	}
	if !fields.Contains(4) {
		r.Region = nil
	}
	if !fields.Contains(8) {
		r.RegionName = nil
	}
	if !fields.Contains(16) {
		r.City = nil
	}
	if !fields.Contains(524288) {
		r.District = nil
	}
	if !fields.Contains(32) {
		r.Zip = nil
	}
	if !fields.Contains(64) {
		r.Lat = nil
	}
	if !fields.Contains(128) {
		r.Lon = nil
	}
	if !fields.Contains(256) {
		r.Timezone = nil
	}
	if !fields.Contains(33554432) {
		r.Offset = nil
	}
	if !fields.Contains(8388608) {
		r.Currency = nil
	}
	if !fields.Contains(512) {
		r.ISP = nil
	}
	if !fields.Contains(1024) {
		r.Org = nil
	}
	if !fields.Contains(2048) {
		r.AS = nil
	}
	if !fields.Contains(4194304) {
		r.ASName = nil
	}
	if !fields.Contains(4096) {
		r.Reverse = nil
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
		r.Message = nil
	}
	if !fields.Contains(8192) {
		r.Query = nil
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
		len(c.Lang)

	if c.Response.Status != nil {
		size += len(*c.Response.Status)
	}
	if c.Response.Continent != nil {
		size += len(*c.Response.Continent)
	}
	if c.Response.ContinentCode != nil {
		size += len(*c.Response.ContinentCode)
	}
	if c.Response.Country != nil {
		size += len(*c.Response.Country)
	}
	if c.Response.CountryCode != nil {
		size += len(*c.Response.CountryCode)
	}
	if c.Response.Region != nil {
		size += len(*c.Response.Region)
	}
	if c.Response.RegionName != nil {
		size += len(*c.Response.RegionName)
	}
	if c.Response.City != nil {
		size += len(*c.Response.City)
	}
	if c.Response.District != nil {
		size += len(*c.Response.District)
	}
	if c.Response.Zip != nil {
		size += len(*c.Response.Zip)
	}
	if c.Response.Timezone != nil {
		size += len(*c.Response.Timezone)
	}
	if c.Response.Currency != nil {
		size += len(*c.Response.Currency)
	}
	if c.Response.ISP != nil {
		size += len(*c.Response.ISP)
	}
	if c.Response.Org != nil {
		size += len(*c.Response.Org)
	}
	if c.Response.AS != nil {
		size += len(*c.Response.AS)
	}
	if c.Response.ASName != nil {
		size += len(*c.Response.ASName)
	}
	if c.Response.Reverse != nil {
		size += len(*c.Response.Reverse)
	}
	if c.Response.Mobile != nil {
		size += 1
	}
	if c.Response.Proxy != nil {
		size += 1
	}
	if c.Response.Hosting != nil {
		size += 1
	}
	if c.Response.Message != nil {
		size += len(*c.Response.Message)
	}
	if c.Response.Query != nil {
		size += len(*c.Response.Query)
	}

	return size
}

//easyjson:json
type CacheEntries []*CacheEntry
