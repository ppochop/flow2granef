package ipfix

import (
	"net/netip"
	"time"

	"github.com/ppochop/flow2granef/flowutils"
	ipproto "github.com/ppochop/flow2granef/ip-proto"
)

type IpfixprobeFlow struct {
	SrcIp                *netip.Addr
	SrcIp4               *netip.Addr `json:"iana:sourceIPv4Address"`
	SrcIp6               *netip.Addr `json:"iana:sourceIPv6Address"`
	SrcPort              uint16      `json:"iana:sourceTransportPort"`
	DestIp               *netip.Addr
	DestIp4              *netip.Addr `json:"iana:destinationIPv4Address"`
	DestIp6              *netip.Addr `json:"iana:destinationIPv6Address"`
	DestPort             uint16      `json:"iana:destinationTransportPort"`
	Proto                string      `json:"iana:protocolIdentifier"`
	IpVersion            uint8       `json:"iana:ipVersion"`
	FlowStartMs          time.Time   `json:"iana:flowStartMicroseconds"`
	FlowEndMs            time.Time   `json:"iana:flowEndMicroseconds"`
	FlowId               uint64      `json:"iana:flowId"`
	Pkts                 uint64      `json:"iana:packetDeltaCount"`
	PktsReverse          uint64      `json:"iana@reverse:packetDeltaCount@reverse"`
	Bytes                uint64      `json:"iana:octetDeltaCount"`
	BytesReverse         uint64      `json:"iana@reverse:octetDeltaCount@reverse"`
	FlowEndReason        int         `json:"iana:flowEndReason"`
	IcmpTypeCode4        *uint16     `json:"iana:icmpTypeCodeIPv4"`
	IcmpTypeCode4Reverse *uint16     `json:"iana@reverse:icmpTypeCodeIPv4@reverse"`
	IcmpTypeCode6        *uint16     `json:"iana:icmpTypeCodeIPv6"`
	IcmpTypeCode6Reverse *uint16     `json:"iana@reverse:icmpTypeCodeIPv6@reverse"`
	Vlan                 *uint16     `json:"iana:vlanId"`
	DnsTransactionID     *uint16     `json:"cesnet:DNSTransactionID"`
	DnsQName             *string     `json:"cesnet:DNSName"`
	DNSQType             *uint       `json:"cesnet:DNSQType"`
	DNSAnswer            *string     `json:"cesnet:DNSRData"`
	DNSTTL               *uint       `json:"cesnet:DNSRRTTL"`
	HTTPUserAgent        *string     `json:"flowmon:httpUserAgent"`
	HTTPUrl              *string     `json:"flowmon:httpUrl"`
	HTTPHost             *string     `json:"flowmon:httpHost"`
	HTTPStatusCode       *uint16     `json:"flowmon:httpStatusCode"`
	HTTPMethod           *string     `json:"flowmon:httpMethod"`
}

func (f *IpfixprobeFlow) GetGranefFlowRec(source string) *flowutils.FlowRec {
	app := "?"
	if f.IsDns() {
		app = "dns"
	}
	shouldSwitchDir := f.IsDnsAnswer()
	srcIp, destIp := f.GetIPs()
	ret := &flowutils.FlowRec{
		OrigIp:      srcIp,
		OrigPort:    f.SrcPort,
		RespIp:      destIp,
		RespPort:    f.DestPort,
		OrigBytes:   f.Bytes,
		RespBytes:   f.BytesReverse,
		OrigPkts:    f.Pkts,
		RespPkts:    f.PktsReverse,
		FlushReason: f.GetIpfixFlushReason(),
		FirstTs:     f.GetFirstTs(),
		LastTs:      f.GetLastTs(),
		Protocol:    f.GetProto(),
		App:         app,
		FlowSource:  source,
	}
	// ipfixprobe treats DNS answers in a uniflow fashion and can even export them earlier than the queries
	if shouldSwitchDir {
		ret.OrigIp = destIp
		ret.OrigPort = f.DestPort
		ret.RespIp = srcIp
		ret.RespPort = f.SrcPort
		ret.OrigBytes, ret.RespBytes = ret.RespBytes, ret.OrigBytes
		ret.OrigPkts, ret.RespPkts = ret.RespPkts, ret.OrigPkts
	}
	return ret
}

func (f *IpfixprobeFlow) GetGranefDNSRec() *flowutils.DNSRec {
	ipAnswer, err := netip.ParseAddr(*f.DNSAnswer)
	if err != nil {
		// Answer not an IP addr, throwing away
		return nil
	}
	ret := &flowutils.DNSRec{
		TransId: f.DnsTransactionID,
		Query:   f.DnsQName,
		Answer:  []*netip.Addr{&ipAnswer},
		TTL:     []*uint{f.DNSTTL},
	}
	if f.DNSQType != nil {
		qtype := ipproto.RRTypeFromNum(uint16(*f.DNSQType))
		qtype_name := qtype.GetName()
		ret.QType = &qtype_name
	}
	return ret
}

func (f *IpfixprobeFlow) GetGranefHTTPRec() *flowutils.HTTPRec {
	switch {
	case *f.HTTPHost == "":
		return nil
	case *f.HTTPUrl == "":
		return nil
	}

	ret := &flowutils.HTTPRec{
		Hostname:  f.HTTPHost,
		Url:       f.HTTPUrl,
		UserAgent: f.HTTPUserAgent,
	}
	if f.HTTPMethod != nil {
		ret.Method = f.HTTPMethod
	}
	if f.HTTPStatusCode != nil {
		ret.StatusCode = *f.HTTPStatusCode
	}
	return ret
}

func (f *IpfixprobeFlow) GetFirstTs() time.Time {
	return f.FlowStartMs.UTC()
}

func (f *IpfixprobeFlow) GetLastTs() time.Time {
	return f.FlowEndMs.UTC()
}

func (f *IpfixprobeFlow) GetIPs() (srcIp *netip.Addr, destIp *netip.Addr) {
	if f.SrcIp6 == nil {
		return f.SrcIp4, f.DestIp4
	} else {
		return f.SrcIp6, f.DestIp6
	}
}

func (f *IpfixprobeFlow) GetProto() ipproto.Protocol {
	return ipproto.ProtocolFromName(f.Proto)
}

func (f *IpfixprobeFlow) IsDns() bool {
	return f.DnsTransactionID != nil
}

func (f *IpfixprobeFlow) IsDnsAnswer() bool {
	return f.DNSAnswer != nil && *f.DNSAnswer != ""
}

func (f *IpfixprobeFlow) GetIpfixFlushReason() flowutils.FlushReason {
	if f.IsDns() {
		// ipfixprobe treats DNS answers in a uniflow fashion and can even export them earlier than the queries
		return flowutils.ActiveTimeout
	}
	switch f.FlowEndReason {
	case 1:
		return flowutils.PassiveTimeout
	case 2, 4, 5:
		return flowutils.ActiveTimeout
	case 3:
		return flowutils.Finished
	default:
		return flowutils.Unknown
	}
}
