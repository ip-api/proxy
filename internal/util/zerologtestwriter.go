package util

import "testing"

type ZerologTestWriter struct {
	T *testing.T
}

func (z ZerologTestWriter) Write(p []byte) (n int, err error) {
	z.T.Log(string(p))
	return len(p), nil
}
