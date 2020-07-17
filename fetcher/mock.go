package fetcher

import (
	"errors"
	"sync"
	"time"

	"github.com/ip-api/proxy/structs"
	"github.com/ip-api/proxy/util"
)

type Mock struct {
	sync.Mutex
	Requests []int // batch size of each batch request.
}

func (mo *Mock) Fetch(m map[string]*structs.CacheEntry) error {
	mo.Lock()
	defer mo.Unlock()

	mo.Requests = append(mo.Requests, len(m))

	if _, ok := m["0.0.0.0en"]; ok {
		return errors.New("test error")
	}

	for key, entry := range m {
		entry.Response = MockResponseFor(key)
		entry.Expires = util.Now().Add(time.Minute)
	}
	return nil
}

func (mo *Mock) FetchSelf(lang string) (structs.Response, error) {
	mo.Lock()
	defer mo.Unlock()

	return structs.Response{}, nil
}

func (mo *Mock) Debug() interface{} {
	return nil
}

func intp(i int) *int {
	return &i
}

func str(s string) *string {
	return &s
}

func flt(f float64) *float64 {
	return &f
}

func boolp(b bool) *bool {
	return &b
}

func MockResponseFor(key string) structs.Response {
	response := structs.Response{
		Status:        str(""),
		Continent:     str(""),
		ContinentCode: str(""),
		Country:       str(""),
		CountryCode:   str(""),
		Region:        str(""),
		RegionName:    str(""),
		City:          str(""),
		District:      str(""),
		Zip:           str(""),
		Lat:           flt(0),
		Lon:           flt(0),
		Timezone:      str(""),
		Offset:        intp(0),
		Currency:      str(""),
		ISP:           str(""),
		Org:           str(""),
		AS:            str(""),
		ASName:        str(""),
		Reverse:       str(""),
		Mobile:        boolp(false),
		Proxy:         boolp(false),
		Hosting:       boolp(false),
		Message:       str(""),
		Query:         str(""),
	}

	switch key {
	case "1.1.1.1en":
		response.Country = str("Some Country")
		response.City = str("Some City")
		response.Query = str("1.1.1.1")
	case "1.1.1.1ja":
		response.Country = str("Some japanese Country")
		response.City = str("Some japanese City")
		response.Query = str("1.1.1.1")
	case "2.2.2.2en":
		response.Status = str("success")
		response.Country = str("Some other Country")
		response.CountryCode = str("SO")
		response.Region = str("SX")
		response.RegionName = str("Some other Region")
		response.City = str("Some other City")
		response.Zip = str("some other zip")
		response.Lat = flt(13)
		response.Lon = flt(37)
		response.Timezone = str("some/timezone")
		response.ISP = str("Some other ISP")
		response.Org = str("Some other Org")
		response.AS = str("Some other AS")
		response.Query = str("2.2.2.2")
	default:
		response.Country = str(key)
		response.City = str(key)
		response.Query = str(key[:len(key)-2])
	}

	return response
}
