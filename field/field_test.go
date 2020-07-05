package field_test

import (
	"testing"

	"github.com/ip-api/cache/field"
)

func TestFromCSV(t *testing.T) {
	f := field.FromCSV("district,isp,query")
	if f != 532992 {
		t.Errorf("expected 532992 got %d", f)
	}
}

func TestContains(t *testing.T) {
	a := field.FromCSV("district,isp,query")
	b := field.FromCSV("isp,district")

	if !a.Contains(b) {
		t.Errorf("%s does not contain %s", a, b)
	}
}

func TestMerge(t *testing.T) {
	a := field.FromCSV("district,isp,query")
	b := field.FromCSV("isp,district,timezone")
	c := a.Merge(b)

	if c != 533248 {
		t.Errorf("expected 533248 got %d", c)
	}

	if !c.Contains(a) {
		t.Errorf("%s does not contain %s", c, a)
	}
	if !c.Contains(b) {
		t.Errorf("%s does not contain %s", c, b)
	}
}
