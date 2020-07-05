package batch

import (
	"sync"
	"time"

	"github.com/ip-api/cache/cache"
	"github.com/ip-api/cache/fetcher"
	"github.com/ip-api/cache/field"
	"github.com/ip-api/cache/structs"
	"github.com/ip-api/cache/wait"
	"github.com/rs/zerolog"
)

const (
	maxBatchEntries = 100
)

type batch struct {
	entries map[string]*structs.CacheEntry
	bae     *wait.BlockAndError
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
			bae: &wait.BlockAndError{
				C: make(chan struct{}),
			},
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
	var running *batch

	b.mu.Lock()
	{
		if len(b.next.entries) == 0 {
			// There are no requests in the batch.
			// No need to create a new batch, just keep using the current one.
			b.mu.Unlock()
			return
		}

		b.running = append(b.running, b.next)
		running = b.next
		b.next = &batch{
			entries: make(map[string]*structs.CacheEntry, len(b.next.entries)),
			bae: &wait.BlockAndError{
				C: make(chan struct{}),
			},
		}
	}
	b.mu.Unlock()

	b.logger.Debug().Msgf("batch with %d entries", len(running.entries))

	err := b.client.Fetch(running.entries)

	b.mu.Lock()
	if err != nil {
		running.bae.Err = err
	} else {
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
	close(running.bae.C)
	b.mu.Unlock()
}

func (b *Batches) Add(ip string, lang string, fields field.Fields) (*structs.CacheEntry, *wait.BlockAndError) {
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
			return entry, r.bae
		}
	}

	entry, ok := b.next.entries[key]
	if ok {
		// Make sure all fields are in the entry.
		entry.Fields = entry.Fields.Merge(fields)

		return entry, b.next.bae
	}

	entry = &structs.CacheEntry{
		IP:     ip,
		Lang:   lang,
		Fields: fields,
	}
	b.next.entries[key] = entry

	if len(b.next.entries) >= maxBatchEntries {
		go b.Process()
	}

	return entry, b.next.bae
}
