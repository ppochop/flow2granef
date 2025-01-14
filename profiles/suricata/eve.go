package suricata

import (
	"net/netip"
	"time"

	"github.com/ppochop/flow2granef/flowutils"
	ipproto "github.com/ppochop/flow2granef/ip-proto"
)

type SuricataEventType string

const (
	SuricataEventFlow    SuricataEventType = "flow"
	SuricataEventDns     SuricataEventType = "dns"
	SuricataEventHttp    SuricataEventType = "http"
	SuricataEventUnknown SuricataEventType = "err"
)

type SuricataEve struct {
	FlowId       uint64            `json:"flow_id"`
	EventType    string            `json:"event_type"`
	SrcIp        netip.Addr        `json:"src_ip"`
	SrcPort      uint16            `json:"src_port"`
	DestIp       netip.Addr        `json:"dest_ip"`
	DestPort     uint16            `json:"dest_port"`
	Proto        string            `json:"proto"`
	AppProto     string            `json:"app_proto"`
	IcmpType     *uint8            `json:"icmp_type"`
	IcmpCode     *uint8            `json:"icmp_code"`
	IcmpRespType *uint8            `json:"response_icmp_type"`
	IcmpRespCode *uint8            `json:"response_icmp_code"`
	Flow         *SuricataFlowInfo `json:"flow"`
	Dns          *SuricataDnsInfo  `json:"dns"`
	Http         *SuricataHttp     `json:"http"`
	Vlan         []uint16          `json:"vlan"`
}

func (s *SuricataEve) DetermineEventType() SuricataEventType {
	switch s.EventType {
	case "flow":
		return SuricataEventFlow
	case "dns":
		return SuricataEventDns
	case "http":
		return SuricataEventHttp
	default:
		return SuricataEventUnknown
	}
}

func (s *SuricataEve) GetGranefFlowRec(source string) *flowutils.FlowRec {
	return &flowutils.FlowRec{
		OrigIp:      &s.SrcIp,
		OrigPort:    s.SrcPort,
		RespIp:      &s.DestIp,
		RespPort:    s.DestPort,
		FlushReason: s.Flow.GetSuricataFlushReason(),
		FirstTs:     s.Flow.GetFirstTs(),
		LastTs:      s.Flow.GetLastTs(),
		Protocol:    ipproto.ProtocolFromName(s.Proto),
		App:         s.AppProto,
		FlowSource:  source,
	}
}

func (s *SuricataEve) GetGranefMiminalFlowRec(source string) *flowutils.FlowRec {
	return &flowutils.FlowRec{
		OrigIp:     &s.SrcIp,
		OrigPort:   s.SrcPort,
		RespIp:     &s.DestIp,
		RespPort:   s.DestPort,
		Protocol:   ipproto.ProtocolFromName(s.Proto),
		FlowSource: source,
	}
}

func (z *SuricataTransformer) GetSuricataTimestamp(ts string) time.Time {
	ts2 := z.reTime.ReplaceAllString(ts, "$1:$2")
	ret := time.Time{}
	ret.UnmarshalText([]byte(ts2))
	return ret
}
