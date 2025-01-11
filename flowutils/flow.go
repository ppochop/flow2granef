package flowutils

import (
	"net/netip"
	"time"

	ipproto "github.com/ppochop/flow2granef/ip-proto"
	"github.com/satta/gommunityid"
)

type FlushReason string

const (
	ActiveTimeout  FlushReason = "a"
	PassiveTimeout FlushReason = "p"
	Finished       FlushReason = "f"
	Unknown        FlushReason = "?"
)

type FlowRec struct {
	OrigIp      *netip.Addr
	OrigPort    uint16
	RespIp      *netip.Addr
	RespPort    uint16
	FlushReason FlushReason
	FirstTs     time.Time
	LastTs      time.Time
	Protocol    ipproto.Protocol
	App         string
	FlowSource  string
}

func (f *FlowRec) GetFlowTuple() gommunityid.FlowTuple {
	ft := gommunityid.MakeFlowTuple(f.OrigIp.AsSlice(), f.RespIp.AsSlice(), f.OrigPort, f.RespPort, f.Protocol.GetNum())
	return ft
}
