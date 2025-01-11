package zeek

import (
	"encoding/binary"
	"net/netip"
)

type ZeekPre struct {
	OrigIp netip.Addr `json:"id.orig_h"`
	RespIp netip.Addr `json:"id.resp_h"`
}

func (p *ZeekPre) GetPreflowId() uint32 {
	ip1 := p.OrigIp.AsSlice()
	ip2 := p.RespIp.AsSlice()
	ip1Num := binary.BigEndian.Uint32(ip1)
	ip2Num := binary.BigEndian.Uint32(ip2)
	return ip1Num + ip2Num
}
