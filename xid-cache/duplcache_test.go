package xidcache

import (
	"testing"
	"time"
)

func TestAddToEmptyCache(t *testing.T) {
	c := NewDuplCache(1 * time.Hour)
	cid := "1:7r1mTaBaonX4v+g9wvCLrmB63ic="
	timeStart := time.Time{}
	timeStart.UnmarshalText([]byte("2013-02-26T22:04:09.904649+00:00"))
	timeEnd := time.Time{}
	timeEnd.UnmarshalText([]byte("2013-02-26T22:07:02.188702+00:00"))
	_, found := c.DuplHandle(cid, timeStart, timeEnd, "suricata")
	if found {
		t.Fatalf("Got a hit on an empty cache.")
	}
}

func TestDuplCacheHit(t *testing.T) {
	// first add Suricata flow
	c := NewDuplCache(1 * time.Hour)
	cid := "1:7r1mTaBaonX4v+g9wvCLrmB63ic="
	timeStart := time.Time{}
	timeStart.UnmarshalText([]byte("2013-02-26T22:04:09.904649+00:00"))
	timeEnd := time.Time{}
	timeEnd.UnmarshalText([]byte("2013-02-26T22:07:02.188702+00:00"))
	_, found := c.DuplHandle(cid, timeStart, timeEnd, "suricata")
	if found {
		t.Fatalf("Got a hit on an empty cache.")
	}

	// add overlapping flow
	cid2 := "1:7r1mTaBaonX4v+g9wvCLrmB63ic="
	timeStart2 := time.Time{}
	timeStart2.UnmarshalText([]byte("2013-02-26T22:04:10.190Z"))
	timeEnd2 := time.Time{}
	timeEnd2.UnmarshalText([]byte("2013-02-26T22:04:18.340Z"))
	_, found = c.DuplHandle(cid2, timeStart2, timeEnd2, "ipfixprobe")
	if !found {
		t.Fatalf("Did not get a hit when expected.")
	}
}

func TestDuplCacheMiss(t *testing.T) {
	// first add one flow record
	c := NewDuplCache(1 * time.Hour)
	cid := "1:7r1mTaBaonX4v+g9wvCLrmB63ic="
	timeStart := time.Time{}
	timeStart.UnmarshalText([]byte("2013-02-26T22:04:09.904Z"))
	timeEnd := time.Time{}
	timeEnd.UnmarshalText([]byte("2013-02-26T22:04:10.190Z"))
	_, found := c.DuplHandle(cid, timeStart, timeEnd, "ipfixprobe")
	if found {
		t.Fatalf("Got a hit on an empty cache.")
	}

	// add subsequent flow record
	cid2 := "1:7r1mTaBaonX4v+g9wvCLrmB63ic="
	timeStart2 := time.Time{}
	timeStart2.UnmarshalText([]byte("2013-02-26T22:04:10.190Z"))
	timeEnd2 := time.Time{}
	timeEnd2.UnmarshalText([]byte("2013-02-26T22:04:18.340Z"))
	_, found = c.DuplHandle(cid2, timeStart2, timeEnd2, "ipfixprobe")
	if found {
		t.Fatalf("Got a hit when shouldn't.")
	}
}
