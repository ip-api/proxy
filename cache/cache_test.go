package cache_test

import (
	"math/rand"
	"strconv"
	"testing"
	"unsafe"

	"github.com/erikdubbelboer/ip-api-proxy/cache"
	"github.com/erikdubbelboer/ip-api-proxy/structs"
)

func TestCache(t *testing.T) {
	c := cache.New(100000)

	emptySize := int(unsafe.Sizeof(structs.CacheEntry{}))

	ip := "1234567890"
	c.Add("a", &structs.CacheEntry{
		IP: ip,
	})

	size := c.Size()
	expectedSize := emptySize + len(ip)
	if size != expectedSize {
		t.Errorf("expected %d got %d", expectedSize, size)
	}

	rand.Seed(0)

	for i := 0; i < 10000; i++ {
		ip := make([]byte, rand.Intn(1000))
		rand.Read(ip)

		c.Add(strconv.Itoa(i), &structs.CacheEntry{
			IP: string(ip),
		})
	}

	size = c.Size()
	expectedSize = 99449
	if size != expectedSize {
		t.Errorf("expected %d got %d", expectedSize, size)
	}

	if c.Get("0") != nil {
		t.Error(`entry "0" should be evicted`)
	}
}
