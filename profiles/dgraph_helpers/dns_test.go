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

func TestDnsBuildAll(t *testing.T) {
	ttl := uint(2)
	ttls := []*uint{&ttl}
	ip1 := netip.MustParseAddr("31.13.73.26")
	answers := []*netip.Addr{&ip1}
	id := uint16(10102)
	q := "sphotos-b.xx.fbcdn.net"
	dnsRec := flowutils.DNSRec{
		TransId: &id,
		Query:   &q,
		Answer:  answers,
		TTL:     ttls,
	}
	req := BuildDnsTxn(&dnsRec, "flowxid", "dnsxid")
	if len(req.Mutations) == 0 {
		t.Fatalf("error")
	}
}
