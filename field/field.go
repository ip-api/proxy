package field

import (
	"strconv"
	"strings"
)

const Default = 61439 // status,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query,message

const (
	FieldReverse = 4096
	FieldStatus  = 16384
)

var fields = map[string]int{
	"country":       1,
	"countryCode":   2,
	"region":        4,
	"regionName":    8,
	"city":          16,
	"zip":           32,
	"lat":           64,
	"lon":           128,
	"timezone":      256,
	"isp":           512,
	"org":           1024,
	"as":            2048,
	"reverse":       4096,
	"query":         8192,
	"status":        16384,
	"message":       32768,
	"mobile":        65536,
	"proxy":         131072,
	"accuracy":      262144,
	"district":      524288,
	"continent":     1048576,
	"continentCode": 2097152,
	"asname":        4194304,
	"currency":      8388608,
	"hosting":       16777216,
	"offset":        33554432,
}

type Fields int

func FromInt(i int) Fields {
	return Fields(i)
}

func FromCSV(s string) Fields {
	i := 0
	for _, field := range strings.Split(s, ",") {
		if n, ok := fields[field]; ok {
			i |= n
		}
	}
	return FromInt(i)
}

func (f Fields) Contains(o Fields) bool {
	return f&o == o
}

func (f Fields) Merge(o Fields) Fields {
	return f | o
}

func (f Fields) String() string {
	parts := make([]string, 0)
	for name, n := range fields {
		if f.Contains(FromInt(n)) {
			parts = append(parts, name)
		}
	}
	return strings.Join(parts, ",")
}

func (f Fields) Num() string {
	return strconv.Itoa(int(f))
}

func (f Fields) Remove(o Fields) Fields {
	return f & ^o
}
