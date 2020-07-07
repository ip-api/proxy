package fetcher

import (
	"net/http"
	"os"
	"time"

	"github.com/ip-api/proxy/structs"
	"github.com/ip-api/proxy/util"
	"github.com/mailru/easyjson/jwriter"
	"github.com/valyala/fasthttp"
)

type Client interface {
	Fetch(map[string]*structs.CacheEntry) error
	FetchSelf(lang string) (structs.Response, error)
}

type ipApi struct {
	client   *fasthttp.Client
	batchURL string
	selfURL  string
	ttl      time.Duration
}

func NewIPApi() (*ipApi, error) {
	client := &fasthttp.Client{
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		MaxConnsPerHost:               100,
		ReadTimeout:                   time.Second,
		WriteTimeout:                  time.Second,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true, // We always set the correct case on our header.
	}

	base := os.Getenv("IP_API_BASE")
	if base == "" {
		base = "https://pro.ip-api.com"
	}

	ttl := time.Minute
	if v := os.Getenv("TTL"); v != "" {
		if d, err := time.ParseDuration(v); err != nil {
			return nil, err
		} else {
			ttl = d
		}
	}

	return &ipApi{
		client:   client,
		batchURL: base + "/batch?key=" + os.Getenv("IP_API_KEY"),
		selfURL:  base + "/json/?key=" + os.Getenv("IP_API_KEY") + "&lang=",
		ttl:      ttl,
	}, nil
}

func (f *ipApi) Fetch(m map[string]*structs.CacheEntry) error {
	entries := make(structs.CacheEntries, 0, len(m))
	for _, entry := range m {
		entries = append(entries, entry)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(f.batchURL)
	req.Header.SetMethod(http.MethodPost)

	jw := &jwriter.Writer{}
	entries.MarshalEasyJSON(jw)
	if _, err := jw.DumpTo(req.BodyWriter()); err != nil {
		return err
	}

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	if err := f.client.Do(req, res); err != nil {
		return err
	}

	var responses structs.Responses
	if err := responses.UnmarshalJSON(res.Body()); err != nil {
		return err
	}

	for i, entry := range entries {
		entry.Response = responses[i]
		entry.Expires = util.Now().Add(f.ttl)
	}

	return nil
}

func (f *ipApi) FetchSelf(lang string) (structs.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(f.selfURL + lang)
	req.Header.SetMethod(http.MethodGet)
	req.Header.SetContentType("application/json")

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	if err := f.client.Do(req, res); err != nil {
		return structs.Response{}, err
	}

	var response structs.Response
	if err := response.UnmarshalJSON(res.Body()); err != nil {
		return structs.Response{}, err
	}

	return response, nil
}
