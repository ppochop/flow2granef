package ipfix

import (
	"encoding/binary"
	"net/netip"
)

type IpfixprobePreFlow struct {
	SrcIp   *netip.Addr
	SrcIp4  *netip.Addr `json:"iana:sourceIPv4Address"`
	SrcIp6  *netip.Addr `json:"iana:sourceIPv6Address"`
	DestIp  *netip.Addr
	DestIp4 *netip.Addr `json:"iana:destinationIPv4Address"`
	DestIp6 *netip.Addr `json:"iana:destinationIPv6Address"`
}

func (p *IpfixprobePreFlow) FixIPs() {
	if p.SrcIp6 == nil {
		p.SrcIp = p.SrcIp4
		p.DestIp = p.DestIp4
	} else {
		p.SrcIp = p.SrcIp6
		p.DestIp = p.DestIp6
	}
}

func (p *IpfixprobePreFlow) GetIpfixPreflowId() uint32 {
	var ip1 []byte
	var ip2 []byte
	if p.SrcIp6 == nil {
		ip1 = p.SrcIp4.AsSlice()
		ip2 = p.DestIp4.AsSlice()
	} else {
		ip1 = p.SrcIp6.AsSlice()
		ip2 = p.DestIp6.AsSlice()
	}
	ip1Num := binary.BigEndian.Uint32(ip1)
	ip2Num := binary.BigEndian.Uint32(ip2)
	return ip1Num + ip2Num
}
