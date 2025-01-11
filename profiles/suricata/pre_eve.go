package suricata

import (
	"encoding/binary"
	"net/netip"
)

type SuricataPreEve struct {
	SrcIp  netip.Addr `json:"src_ip"`
	DestIp netip.Addr `json:"dest_ip"`
}

func (p *SuricataPreEve) GetPreflowId() uint32 {
	ip1 := p.SrcIp.AsSlice()
	ip2 := p.DestIp.AsSlice()
	ip1Num := binary.BigEndian.Uint32(ip1)
	ip2Num := binary.BigEndian.Uint32(ip2)
	return ip1Num + ip2Num
}
