package xidcache

import (
	"testing"
	"time"
)

func TestOrdered(t *testing.T) {
	c := New(10 * time.Minute)
	commId := "1:DQJeVOCkTS9+fMa+rYPZ5vXu51A="
	flowIdQ := "1046060674446557591"
	ftsQ := time.Time{}
	ftsQ.UnmarshalText([]byte("2013-02-26T22:02:56.271Z"))
	atsQ := ftsQ
	flowIdA := "13220912849057427144"
	ftsA := time.Time{}
	ftsA.UnmarshalText([]byte("2013-02-26T22:02:56.293Z"))
	atsA := ftsA

	var hit CacheHitResult
	xid1, hit := c.AddOrGet(commId, false, flowIdQ, ftsQ, atsQ)
	if hit == Hit {
		t.Fatalf("hit empty cache")
	}
	if xid1 != flowIdQ {
		t.Fatalf("unexpected returned xid")
	}

	xid2, hit := c.AddOrGet(commId, false, flowIdA, ftsA, atsA)
	if hit == Miss {
		t.Fatalf("did not hit cache when should")
	}
	if xid2 != flowIdQ {
		t.Fatalf("should have returned original xid")
	}
}
