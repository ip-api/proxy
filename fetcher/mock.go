package fetcher

import (
	"time"

	"github.com/ip-api/cache/structs"
	"github.com/ip-api/cache/util"
)

type Mock struct {
	Requests []int // batch size of each batch request.
}

func (mo *Mock) Fetch(m map[string]*structs.CacheEntry) error {
	mo.Requests = append(mo.Requests, len(m))

	for key, entry := range m {
		entry.Response = MockResponseFor(key)
		entry.Expires = util.Now().Add(time.Minute)
	}
	return nil
}

func (mo *Mock) FetchSelf(lang string) (structs.Response, error) {
	return structs.Response{}, nil
}

func MockResponseFor(key string) structs.Response {
	switch key {
	case "1.1.1.1en":
		return structs.Response{
			Country: "Some Country",
			City:    "Some City",
			Query:   "1.1.1.1",
		}
	case "1.1.1.1ja":
		return structs.Response{
			Country: "Some japanese Country",
			City:    "Some japanese City",
			Query:   "1.1.1.1",
		}
	case "2.2.2.2en":
		// status,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query
		return structs.Response{
			Status:      "success",
			Country:     "Some other Country",
			CountryCode: "SO",
			Region:      "SX",
			RegionName:  "Some other Region",
			City:        "Some other City",
			Zip:         "some other zip",
			Lat:         13,
			Lon:         37,
			Timezone:    "some/timezone",
			ISP:         "Some other ISP",
			Org:         "Some other Org",
			AS:          "Some other AS",
			Query:       "2.2.2.2",
		}
	default:
		return structs.Response{
			Country: key,
			City:    key,
			Query:   key[:len(key)-2],
		}
	}
}
