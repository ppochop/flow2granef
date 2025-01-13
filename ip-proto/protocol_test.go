package ipproto

import (
	"testing"
)

func TestIpProto(t *testing.T) {
	protoStr := "tcp"
	proto := ProtocolFromName(protoStr)
	if proto.GetNum() != 6 {
		t.Fatalf("invalid tcp number")
	}
}

func TestDnsRRtype(t *testing.T) {
	rrtypeStr := "aaaa"
	rrType := RRTypeFromName(rrtypeStr)
	if rrTypeNum := rrType.GetNum(); rrTypeNum != 28 {
		t.Fatalf("invalid aaaa number %d", rrTypeNum)
	}
}
