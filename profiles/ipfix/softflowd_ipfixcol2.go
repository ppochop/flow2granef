package ipfix

import (
	"net/netip"
	"time"

	"github.com/ppochop/flow-normalize/flow"
	"github.com/ppochop/flow-normalize/profiles"
	"github.com/satta/gommunityid"
)

type SoftflowdFlow struct {
	SrcIp4               netip.Addr `json:"iana:sourceIPv4Address"`
	SrcIp6               netip.Addr `json:"iana:sourceIPv6Address"`
	SrcPort              uint16     `json:"iana:sourceTransportPort"`
	DestIp4              netip.Addr `json:"iana:destinationIPv4Address"`
	DestIp6              netip.Addr `json:"iana:destinationIPv6Address"`
	DestPort             uint16     `json:"iana:destinationTransportPort"`
	Proto                string     `json:"iana:protocolIdentifier"`
	IpVersion            uint8      `json:"iana:ipVersion"`
	FlowStartMs          time.Time  `json:"iana:flowStartMilliseconds"`
	FlowEndMs            time.Time  `json:"iana:flowEndMilliseconds"`
	Pkts                 uint       `json:"iana:packetDeltaCount"`
	PktsReverse          uint       `json:"iana@reverse:packetDeltaCount@reverse"`
	Bytes                uint       `json:"iana:octetDeltaCount"`
	BytesReverse         uint       `json:"iana@reverse:octetDeltaCount@reverse"`
	FlowEndReason        int        `json:"iana:flowEndReason"`
	IcmpTypeCode4        uint16     `json:"iana:icmpTypeCodeIPv4"`
	IcmpTypeCode4Reverse uint16     `json:"iana@reverse:icmpTypeCodeIPv4@reverse"`
	IcmpTypeCode6        uint16     `json:"iana:icmpTypeCodeIPv6"`
	IcmpTypeCode6Reverse uint16     `json:"iana@reverse:icmpTypeCodeIPv6@reverse"`
	Vlan                 uint16     `json:"iana:vlanId"`
}

type SoftflowdTransformer struct {
	commIdGen gommunityid.CommunityID
}

func init() {
	profiles.RegisterTransformer("softflowd", InitSoftflowdTransformer)
}

func InitSoftflowdTransformer() profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &SoftflowdTransformer{
		commIdGen: commId,
	}
}

func (z *SoftflowdTransformer) GetProfile() any {
	return &SoftflowdFlow{}
}

func (z *SoftflowdTransformer) ToFlow(sFAny any) (*flow.Flow, *flow.Flow, error) {
	sF := sFAny.(*SoftflowdFlow)

	dur := sF.FlowEndMs.Sub(sF.FlowStartMs)
	proto := flow.ProtocolFromName(sF.Proto)

	var icmpType, icmpCode uint8
	if proto.IsIcmp() {
		icmpType, icmpCode = handleSoftflowdIcmp(sF)
	}

	var srcIp, destIp netip.Addr
	if sF.IpVersion == 4 { // IPv4
		srcIp = sF.SrcIp4
		destIp = sF.DestIp4
	} else {
		srcIp = sF.SrcIp6
		destIp = sF.DestIp6
	}

	var ft gommunityid.FlowTuple
	if proto.IsIcmp() {
		ft = gommunityid.MakeFlowTuple(srcIp.AsSlice(), destIp.AsSlice(), uint16(icmpType), uint16(icmpCode), proto.GetNum())
	} else {
		ft = gommunityid.MakeFlowTuple(srcIp.AsSlice(), destIp.AsSlice(), sF.SrcPort, sF.DestPort, proto.GetNum())
	}

	commId := z.commIdGen.CalcBase64(ft)
	return &flow.Flow{
		OriginatorIp:   srcIp,
		OriginatorPort: sF.SrcPort,
		ResponderIp:    destIp,
		ResponderPort:  sF.DestPort,
		FirstTs:        sF.FlowStartMs.UTC(),
		LastTs:         sF.FlowEndMs.UTC(),
		Duration:       dur,
		Protocol:       proto,
		Application:    "unknown",
		FlushReason:    _GetIpfixFlushReason(sF.FlowEndReason),
		FromOrigPkts:   sF.Pkts,
		FromOrigBytes:  sF.Bytes,
		FromRespPkts:   sF.PktsReverse,
		FromRespBytes:  sF.BytesReverse,
		TotalPkts:      sF.Pkts + sF.PktsReverse,
		TotalBytes:     sF.Bytes + sF.BytesReverse,
		Vlan:           sF.Vlan,
		IcmpMsgType:    icmpType,
		IcmpMsgCode:    icmpCode,
		CommunityId:    commId,
		FlowSource:     "softflowd_ipfixcol2",
	}, nil, nil
}

func _GetIpfixFlushReason(flowEndReason int) flow.FlushReason {
	switch flowEndReason {
	case 1:
		return flow.PassiveTimeout
	case 2, 4, 5:
		return flow.ActiveTimeout
	case 3:
		return flow.Finished
	default:
		return flow.Unknown
	}
}

func GetIpfixIcmpTypeCode(typecode uint16) (uint8, uint8) {
	icmpType := typecode / 256
	icmpCode := typecode % 256
	return uint8(icmpType), uint8(icmpCode)
}

// In case the flow is ICMP, we have to perform additional actions:
//   - extract the right ICMP type and code
//   - swap the direction of the flow if it's a "response"
//
// The second is because of softflowd's handling of ICMP's response
// where the originator and responder are based on the original message
// rather than the response.
func handleSoftflowdIcmp(sF *SoftflowdFlow) (uint8, uint8) {
	var icmpType, icmpCode uint8

	if sF.Pkts == 0 { // ICMP response
		if sF.IpVersion == 4 {
			icmpType, icmpCode = GetIpfixIcmpTypeCode(sF.IcmpTypeCode4Reverse)
		} else {
			icmpType, icmpCode = GetIpfixIcmpTypeCode(sF.IcmpTypeCode4)
		}
		swapSoftflowdDirection(sF)
	} else {
		if sF.IpVersion == 4 {
			icmpType, icmpCode = GetIpfixIcmpTypeCode(sF.IcmpTypeCode6Reverse)
		} else {
			icmpType, icmpCode = GetIpfixIcmpTypeCode(sF.IcmpTypeCode6)
		}
	}

	return icmpType, icmpCode
}

func swapSoftflowdDirection(sF *SoftflowdFlow) {
	sF.SrcIp4, sF.DestIp4 = sF.DestIp4, sF.SrcIp4
	sF.SrcIp6, sF.DestIp6 = sF.DestIp6, sF.SrcIp6

	sF.SrcPort, sF.DestPort = sF.DestPort, sF.SrcPort

	sF.Pkts, sF.PktsReverse = sF.PktsReverse, sF.Pkts
	sF.Bytes, sF.BytesReverse = sF.BytesReverse, sF.Bytes
}
