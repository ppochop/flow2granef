package flowutils

import "net/netip"

type DNSRec struct {
	TransId *uint16
	Query   *string
	Answer  []*netip.Addr
	TTL     []*uint
	QType   *string
}
