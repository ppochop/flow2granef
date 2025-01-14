package dgraphhelpers

import (
	"net/netip"
	"testing"

	"github.com/ppochop/flow2granef/flowutils"
)

func TestDnsBuild(t *testing.T) {
	ttls := uint(100)
	id := uint16(10102)
	q := "asdf.cm"
	a, _ := netip.ParseAddr("192.168.11.11")
	d := &flowutils.DNSRec{
		TransId: &id,
		Query:   &q,
		Answer:  []*netip.Addr{&a},
		TTL:     []*uint{&ttls},
	}
	ret := BuildDnsTxn(d, "test1", "testdns")
	if len(ret.Mutations) == 0 {
		t.Fatalf("error")
	}
}
