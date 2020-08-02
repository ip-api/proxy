package reverse

import (
	"context"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Reverser interface {
	Lookup(ip string, out *string, wg *sync.WaitGroup)
}

type single struct {
	ip  string
	out *string
	wg  *sync.WaitGroup
}

type reverser struct {
	logger zerolog.Logger

	resolver net.Resolver
	queue    chan single
}

func New(logger zerolog.Logger) Reverser {
	workers := 10
	if v := os.Getenv("REVERSE_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err != nil {
			logger.Fatal().Err(err).Msg("invalid REVERSE_WORKERS")
		} else {
			workers = n
		}
	}

	preferGo := true
	if os.Getenv("REVERSE_PREFERGO") == "false" {
		preferGo = false
	}

	r := &reverser{
		logger: logger,
		resolver: net.Resolver{
			PreferGo: preferGo,
		},
		queue: make(chan single, workers*10),
	}

	for i := 0; i < workers; i++ {
		go r.worker()
	}

	return r
}

func (l *reverser) worker() {
	for s := range l.queue {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
		addrs, err := l.resolver.LookupAddr(ctx, s.ip)
		cancel()
		if err != nil {
			l.logger.Debug().Err(err).Str("ip", s.ip).Msg("failed to do reverse lookup")
			*s.out = ""
		} else {
			if len(addrs) == 0 || len(addrs[0]) == 0 {
				*s.out = ""
			} else {
				a := addrs[0]
				*s.out = a[:len(a)-1]
			}
		}

		s.wg.Done()
	}
}

func (l *reverser) Lookup(ip string, out *string, wg *sync.WaitGroup) {
	wg.Add(1)
	l.queue <- single{
		ip:  ip,
		out: out,
		wg:  wg,
	}
}
