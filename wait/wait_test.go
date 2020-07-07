package wait_test

import (
	"testing"
	"time"

	"github.com/ip-api/cache/wait"
)

func TestWait(t *testing.T) {
	one := make(chan struct{})
	two := make(chan struct{})

	w := wait.New()
	w.Add(one)
	w.Add(one)
	w.Add(two)

	go func() {
		time.Sleep(time.Millisecond * 10)
		close(one)
		time.Sleep(time.Millisecond * 10)
		close(two)
	}()

	start := time.Now()
	w.Wait()
	d := time.Since(start)

	if d < time.Millisecond*20 {
		t.Error("wait was too short")
	}
}
