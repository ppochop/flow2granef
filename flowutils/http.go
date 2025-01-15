package flowutils

import "net/netip"

type HTTPRec struct {
	ClientIp  *netip.Addr
	ServerIp  *netip.Addr
	Url       *string
	Hostname  *string
	UserAgent *string
}
