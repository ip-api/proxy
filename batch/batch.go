package batch

import (
	"sync"
	"time"

	"github.com/ip-api/proxy/cache"
	"github.com/ip-api/proxy/fetcher"
	"github.com/ip-api/proxy/field"
	"github.com/ip-api/proxy/structs"
	"github.com/rs/zerolog"
)

const (
	maxBatchEntries = 100
)

type batch struct {
	entries map[string]*structs.CacheEntry
	c       chan struct{}
}

type Batches struct {
	mu sync.Mutex

	next    *batch
	running []*batch

	logger zerolog.Logger
	cache  *cache.Cache
	client fetcher.Client
}

func New(logger zerolog.Logger, cache *cache.Cache, client fetcher.Client) *Batches {
	return &Batches{
		next: &batch{
			entries: make(map[string]*structs.CacheEntry),
			c:       make(chan struct{}),
		},
		running: make([]*batch, 0),
		logger:  logger,
		cache:   cache,
		client:  client,
	}
}

func (b *Batches) ProcessLoop() {
	for {
		time.Sleep(time.Millisecond * 10)

		b.Process()
	}
}

func (b *Batches) Process() {
	b.mu.Lock()
	b.processLocked()
	b.mu.Unlock()
}

// processLocked assumes b.mu is already locked.
func (b *Batches) processLocked() {
	var running *batch

	if len(b.next.entries) == 0 {
		// There are no requests in the batch.
		// No need to create a new batch, just keep using the current one.
		return
	}

	b.running = append(b.running, b.next)
	running = b.next
	b.next = &batch{
		entries: make(map[string]*structs.CacheEntry, len(b.next.entries)),
		c:       make(chan struct{}),
	}

	b.logger.Debug().Msgf("batch with %d entries", len(running.entries))

	// Fetch multiple batches in parallel in goroutines.
	go func() {
		err := b.client.Fetch(running.entries)

		if err != nil {
			b.logger.Error().Err(err).Msg("error in upstream")
		}

		b.mu.Lock()
		{
			if err == nil {
				for key, entry := range running.entries {
					b.cache.Add(key, entry)
				}
			}

			for i, n := range b.running {
				if n == running {
					b.running[i] = b.running[len(b.running)-1]
					b.running[len(b.running)-1] = nil
					b.running = b.running[:len(b.running)-1]
					break
				}
			}

			close(running.c)
		}
		b.mu.Unlock()
	}()
}

func (b *Batches) Add(ip string, lang string, fields field.Fields) (*structs.CacheEntry, chan struct{}) {
	key := ip + lang

	b.mu.Lock()
	defer b.mu.Unlock()

	entry := b.cache.Get(key)
	if entry != nil {
		// Does the cached entry contain all the fields we need to return?
		if entry.Fields.Contains(fields) {
			return entry, nil
		}
	}

	// Check if the requested fields and IP are already in an outgoing batch request.
	for _, r := range b.running {
		entry, ok := r.entries[key]
		if ok && entry.Fields.Contains(fields) {
			return entry, r.c
		}
	}

	entry, ok := b.next.entries[key]
	if ok {
		// Make sure all fields are in the entry.
		entry.Fields = entry.Fields.Merge(fields)

		return entry, b.next.c
	}

	entry = &structs.CacheEntry{
		IP:       ip,
		Lang:     lang,
		Fields:   fields,
		Response: structs.ErrorResponse("fail", "error in upstream"),
	}
	b.next.entries[key] = entry

	c := b.next.c

	if len(b.next.entries) >= maxBatchEntries {
		b.processLocked()
	}

	return entry, c
}
