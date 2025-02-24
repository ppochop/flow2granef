package zeek

import (
	"encoding/json"
	"net/netip"
	"time"

	"github.com/ppochop/flow2granef/flowutils"
	ipproto "github.com/ppochop/flow2granef/ip-proto"
)

type ZeekConn struct {
	Ts        float64    `json:"ts"`
	Uid       string     `json:"uid"`
	OrigIp    netip.Addr `json:"id.orig_h"`
	OrigPort  uint16     `json:"id.orig_p"`
	RespIp    netip.Addr `json:"id.resp_h"`
	RespPort  uint16     `json:"id.resp_p"`
	Proto     string     `json:"proto"`
	Service   string     `json:"service"`
	Duration  float64    `json:"duration"`
	OrigPkts  uint64     `json:"orig_pkts"`
	OrigBytes uint64     `json:"orig_ip_bytes"`
	RespPkts  uint64     `json:"resp_pkts"`
	RespBytes uint64     `json:"resp_ip_bytes"`
	ConnState string     `json:"conn_state"`
}

type ZeekConnLimited struct {
	Ts       float64    `json:"ts"`
	Uid      string     `json:"uid"`
	OrigIp   netip.Addr `json:"id.orig_h"`
	OrigPort uint16     `json:"id.orig_p"`
	RespIp   netip.Addr `json:"id.resp_h"`
	RespPort uint16     `json:"id.resp_p"`
	Proto    string     `json:"proto"`
	Service  string     `json:"service"`
}

func (z *ZeekTransformer) ZeekHandleConn(data []byte) error {
	zeekConn := ZeekConn{}
	err := json.Unmarshal(data, &zeekConn)
	if err != nil {
		return err
	}
	// send to granef
	return nil
}

func (z *ZeekConn) GetGranefFlowRec(source string) *flowutils.FlowRec {
	return &flowutils.FlowRec{
		OrigIp:      &z.OrigIp,
		OrigPort:    z.OrigPort,
		RespIp:      &z.RespIp,
		RespPort:    z.RespPort,
		OrigBytes:   z.OrigBytes,
		RespBytes:   z.RespBytes,
		OrigPkts:    z.OrigPkts,
		RespPkts:    z.RespPkts,
		FlushReason: z.GetZeekFlushReason(),
		FirstTs:     z.GetFirstTs(),
		LastTs:      z.GetLastTs(),
		Protocol:    ipproto.ProtocolFromName(z.Proto),
		App:         z.Service,
		FlowSource:  source,
	}
}

func (z *ZeekConnLimited) GetGranefFlowRec(source string) *flowutils.FlowRec {
	return &flowutils.FlowRec{
		OrigIp:     &z.OrigIp,
		OrigPort:   z.OrigPort,
		RespIp:     &z.RespIp,
		RespPort:   z.RespPort,
		Protocol:   ipproto.ProtocolFromName(z.Proto),
		App:        z.Service,
		FlowSource: source,
		FirstTs:    z.GetFirstTs(),
		LastTs:     z.GetFirstTs(),
	}
}

func (z *ZeekConn) GetFirstTs() time.Time {
	return time.UnixMilli(int64(z.Ts * 1000)).UTC()
}

func (z *ZeekConnLimited) GetFirstTs() time.Time {
	return time.UnixMilli(int64(z.Ts * 1000)).UTC()
}

func (z *ZeekConn) GetLastTs() time.Time {
	dur := time.Duration(z.Duration * float64(time.Second))
	return z.GetFirstTs().Add(dur).UTC()
}

func (z *ZeekConn) GetZeekFlushReason() flowutils.FlushReason {
	switch z.ConnState {
	case "S0", "S1", "OTH":
		return flowutils.PassiveTimeout
	case "SF", "REJ", "S2", "S3", "RSTO", "RSTR", "RSTOS0", "RSTRH", "SH", "SHR":
		return flowutils.Finished
	default:
		return flowutils.Unknown
	}
}
