package wait_test

import (
	"testing"
	"time"

	"github.com/ip-api/cache/wait"
)

func TestWait(t *testing.T) {
	one := &wait.BlockAndError{
		C: make(chan struct{}),
	}
	two := &wait.BlockAndError{
		C: make(chan struct{}),
	}

	w := wait.New()
	w.Add(one)
	w.Add(one)
	w.Add(two)

	go func() {
		time.Sleep(time.Millisecond * 10)
		close(one.C)
		time.Sleep(time.Millisecond * 10)
		close(two.C)
	}()

	start := time.Now()
	if err := w.Wait(); err != nil {
		t.Fatal(err)
	}
	d := time.Since(start)

	if d < time.Millisecond*20 {
		t.Error("wait was too short")
	}
}
