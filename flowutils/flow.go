// package flowutils provides type definitions for internal representation of supported event types.
package flowutils

import (
	"net/netip"
	"time"

	ipproto "github.com/ppochop/flow2granef/ip-proto"
	"github.com/satta/gommunityid"
)

type FlowRec struct {
	CommId      string
	OrigIp      *netip.Addr
	OrigPort    uint16
	RespIp      *netip.Addr
	RespPort    uint16
	OrigBytes   uint64
	RespBytes   uint64
	OrigPkts    uint64
	RespPkts    uint64
	FlushReason FlushReason
	FirstTs     time.Time
	LastTs      time.Time
	Protocol    ipproto.Protocol
	App         string
	FlowSource  string
}

func (f *FlowRec) GetFlowTuple() gommunityid.FlowTuple {
	ft := gommunityid.MakeFlowTuple(f.OrigIp.AsSlice(), f.RespIp.AsSlice(), f.OrigPort, f.RespPort, uint8(f.Protocol.GetNum())).InOrder()
	return ft
}
