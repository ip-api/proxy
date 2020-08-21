package fetcher

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mailru/easyjson/jwriter"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	"github.com/ip-api/proxy/internal/field"
	"github.com/ip-api/proxy/internal/reverse"
	"github.com/ip-api/proxy/internal/structs"
	"github.com/ip-api/proxy/internal/util"
)

type Client interface {
	Fetch(map[string]*structs.CacheEntry) error
	FetchSelf(lang string, fields field.Fields) (structs.Response, error)
	Debug() interface{}
}

type ipApi struct {
	mu sync.Mutex

	reverser reverse.Reverser

	clients  map[string]*fasthttp.HostClient
	batchURL string
	selfURL  string
	ttl      time.Duration

	servers []*server
	retries int
}

var ErrRetryLimitReached = errors.New("reached retry limit")

func NewIPApi(logger zerolog.Logger, reverser reverse.Reverser) (*ipApi, error) {
	ttl := time.Hour * 24
	if v := os.Getenv("CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err != nil {
			return nil, err
		} else {
			ttl = d
		}
	}

	retries := 4
	if v := os.Getenv("RETRIES"); v != "" {
		if i, err := strconv.Atoi(v); err != nil {
			return nil, err
		} else {
			retries = i
		}
	}

	f := &ipApi{
		reverser: reverser,
		clients:  make(map[string]*fasthttp.HostClient),
		batchURL: "https://pro.ip-api.com/batch?key=" + os.Getenv("IP_API_KEY"),
		selfURL:  "https://pro.ip-api.com/json/?key=" + os.Getenv("IP_API_KEY"),
		ttl:      ttl,
		retries:  retries,
	}

	serverRefreshRate := time.Hour
	if v := os.Getenv("POPS_REFRESH"); v != "" {
		if d, err := time.ParseDuration(v); err != nil {
			logger.Error().Err(err).Msg("invalid POPS_REFRESH")
		} else {
			serverRefreshRate = d
		}
	}

	go func() {
		for {
			servers, err := getServers(logger, f.servers)
			if err != nil {
				logger.Error().Err(err).Msg("failed to fetch pops")

				// Try again after a minute.
				time.Sleep(time.Minute)
				continue
			}

			f.mu.Lock()
			f.servers = servers
			f.mu.Unlock()

			time.Sleep(serverRefreshRate)
		}
	}()

	return f, nil
}

func (f *ipApi) Debug() interface{} {
	return f.servers
}

func (f *ipApi) getBatchServerAndClient() (*server, *fasthttp.HostClient) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Only try servers which haven't return any error in the last minute.
	noErrorAfter := time.Now().Add(-time.Minute)
	var s *server
	for _, ss := range f.servers {
		if !ss.LastError.After(noErrorAfter) {
			s = ss
			break
		}
	}

	// If no server was found we fall back on normal DNS.
	host := "pro.ip-api.com"
	if s != nil {
		host = s.IP
	}

	client, ok := f.clients[host]
	if !ok {
		client = &fasthttp.HostClient{
			Addr:                          "pro.ip-api.com:443",
			IsTLS:                         true,
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			MaxConns:                      100,
			ReadTimeout:                   time.Second,
			WriteTimeout:                  time.Second,
			MaxIdleConnDuration:           time.Minute,
			DisableHeaderNamesNormalizing: true, // We always set the correct case on our header.
			Dial: func(addr string) (net.Conn, error) {
				return fasthttp.Dial(host + ":443")
			},
		}
		f.clients[host] = client
	}

	return s, client
}

func (f *ipApi) Fetch(m map[string]*structs.CacheEntry) error {
	entries := make(structs.CacheEntries, 0, len(m))
	reverses := make([]*string, 0, len(m))

	var wg sync.WaitGroup
	defer wg.Wait() // Wait for all reverse lookups to be done before we return.

	for _, entry := range m {
		entry.Fields = entry.Fields.Merge(field.FieldStatus)

		entries = append(entries, entry)

		if entry.Fields.Contains(field.FieldReverse) {
			s := ""
			reverses = append(reverses, &s)
			f.reverser.Lookup(entry.IP, &s, &wg)
			entry.Fields = entry.Fields.Remove(field.FieldReverse) // Don't also let the backend do a reverse lookup.
		} else {
			reverses = append(reverses, nil)
		}
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	if err := req.URI().Parse(nil, []byte(f.batchURL)); err != nil {
		return err
	}
	req.Header.SetMethod(fasthttp.MethodPost)

	jw := &jwriter.Writer{}
	entries.MarshalEasyJSON(jw)
	if _, err := jw.DumpTo(req.BodyWriter()); err != nil {
		return err
	}

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	var responses structs.Responses
	var err error

	for i := 0; i < f.retries; i++ {
		server, client := f.getBatchServerAndClient()

		if server != nil {
			atomic.AddInt64(&server.Requests, 1)
		}

		if err = client.Do(req, res); err == nil {
			if err = responses.UnmarshalJSON(res.Body()); err == nil {
				if len(responses) != len(entries) {
					if len(responses) == 1 && responses[0].Message != nil {
						return fmt.Errorf("%s", *responses[0].Message)
					}
					return fmt.Errorf("backend response count (%d) doesn't match requested count (%d)", len(responses), len(entries))
				}

				for i, entry := range entries {
					entry.Response = responses[i]
					entry.Expires = util.Now().Add(f.ttl)

					if r := reverses[i]; r != nil {
						entry.Fields = entry.Fields.Merge(field.FieldReverse)

						if entry.Response.Status == nil || *entry.Response.Status != "fail" {
							entry.Response.Reverse = r
						}
					}
				}

				return nil
			}
		}

		if server != nil {
			atomic.AddInt64(&server.Errors, 1)

			f.mu.Lock()
			server.LastError = time.Now()
			f.mu.Unlock()
		}
	}

	if err == nil {
		err = ErrRetryLimitReached
	}

	return err
}

func (f *ipApi) FetchSelf(lang string, fields field.Fields) (structs.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	if err := req.URI().Parse(nil, []byte(f.selfURL+"&lang="+lang+"&fields="+fields.Num())); err != nil {
		return structs.Response{}, err
	}
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetContentType("application/json")

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	var err error

	for i := 0; i < f.retries; i++ {
		server, client := f.getBatchServerAndClient()

		if server != nil {
			atomic.AddInt64(&server.Requests, 1)
		}

		if err = client.Do(req, res); err == nil {
			var response structs.Response
			if err := response.UnmarshalJSON(res.Body()); err == nil {
				return response, nil
			}
		}

		if server != nil {
			atomic.AddInt64(&server.Errors, 1)

			f.mu.Lock()
			server.LastError = time.Now()
			f.mu.Unlock()
		}
	}

	if err == nil {
		err = ErrRetryLimitReached
	}

	return structs.Response{}, err
}
