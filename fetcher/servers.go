package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type server struct {
	IP        string        `json:"ip"`
	Pop       string        `json:"pop"`
	Latency   time.Duration `json:"latency"`
	LastError time.Time     `json:"last_error"`
}

const latencyPings = 4

// latency returns the latency to 'ip' by performing
// latencyPings requests and returning the average latency.
func latency(client *http.Client, ip string) time.Duration {
	u := "http://" + ip + "/ping"

	var measures []time.Duration

	for i := 0; i < latencyPings; i++ {
		start := time.Now()
		if res, err := client.Get(u); err != nil {
			continue
		} else if res.StatusCode != http.StatusOK {
			continue
		}
		measures = append(measures, time.Since(start))
	}

	if len(measures) == 0 {
		return time.Hour
	}

	var measure time.Duration
	for _, m := range measures {
		measure += m
	}

	return measure / time.Duration(len(measures))
}

func getServers(logger zerolog.Logger) ([]*server, error) {
	client := &http.Client{}

	popsUrl := os.Getenv("POPS_URL")
	if popsUrl == "" {
		popsUrl = "https://d2e7s0viy93a0y.cloudfront.net/pops.json"
	}

	res, err := client.Get(popsUrl)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pops endpoint returned: %s", res.Status)
	}

	var servers []*server
	if err := json.NewDecoder(res.Body).Decode(&servers); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(servers))
	for i := range servers {
		go func(i int) {
			defer wg.Done()
			servers[i].Latency = latency(client, servers[i].IP)

			logger.Debug().Dur("latency", servers[i].Latency).Str("ip", servers[i].IP).Msg("latency")
		}(i)
	}
	wg.Wait()

	// We don't need this clients or its connections anymore, so close them.
	client.CloseIdleConnections()

	// Sort the servers by latency ascending.
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Latency < servers[j].Latency
	})

	return servers, nil
}
