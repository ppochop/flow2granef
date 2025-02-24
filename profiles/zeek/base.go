package zeek

type ZeekLogType string

const (
	ZeekLogConn    ZeekLogType = "conn"
	ZeekLogDns     ZeekLogType = "dns"
	ZeekLogHttp    ZeekLogType = "http"
	ZeekLogUnknown ZeekLogType = "err"
)

type ZeekBase struct {
	Uid string `json:"uid"`
	// Conn.log determinant
	OrigPkts *int `json:"orig_pkts"`
	// Http.log determinant
	Method     *string `json:"method"`
	StatusCode *int    `json:"status_code"`
	// Dns.log determinant
	TransId *uint16 `json:"trans_id"`
}

func (z *ZeekBase) decideType() ZeekLogType {
	switch {
	case z.OrigPkts != nil:
		return ZeekLogConn
	case z.Method != nil || z.StatusCode != nil:
		return ZeekLogHttp
	case z.TransId != nil:
		return ZeekLogDns
	default:
		return ZeekLogUnknown
	}
}
