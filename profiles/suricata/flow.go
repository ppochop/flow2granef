package suricata

import (
	"regexp"
	"time"

	"github.com/ppochop/flow2granef/flowutils"
)

type SuricataFlowBypassedInfo struct {
	PktsToServer  uint64 `json:"pkts_toserver"`
	PktsToClient  uint64 `json:"pkts_toclient"`
	BytesToServer uint64 `json:"bytes_toserver"`
	BytesToClient uint64 `json:"bytes_toclient"`
}

type SuricataFlowInfo struct {
	PktsToServer  uint64                    `json:"pkts_toserver"`
	PktsToClient  uint64                    `json:"pkts_toclient"`
	BytesToServer uint64                    `json:"bytes_toserver"`
	BytesToClient uint64                    `json:"bytes_toclient"`
	Start         SuriTime                  `json:"start"`
	End           SuriTime                  `json:"end"`
	Age           uint                      `json:"age"`
	Bypassed      *SuricataFlowBypassedInfo `json:"bypassed"`
	Reason        string                    `json:"reason"`
}

type SuriTime struct {
	time time.Time
}

var reTime = regexp.MustCompile(`"(.+)([\+\-]\d{2})(\d{2})"`)
var reReplace = []byte("$1$2:$3")

func (t *SuriTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	ts := reTime.ReplaceAll(data, reReplace)
	t.time.UnmarshalText(ts)
	return nil
}

func (s *SuricataFlowInfo) GetFirstTs() time.Time {
	return s.Start.time.UTC()
}

func (s *SuricataFlowInfo) GetLastTs() time.Time {
	return s.End.time.UTC()
}

func (s *SuricataFlowInfo) GetSuricataFlushReason() flowutils.FlushReason {
	switch s.Reason {
	case "timeout":
		return flowutils.PassiveTimeout
	case "forced", "shutdown":
		return flowutils.ActiveTimeout
	default:
		return flowutils.Unknown
	}
}
