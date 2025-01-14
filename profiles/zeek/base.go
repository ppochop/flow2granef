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
	// Conn.log stuff
	//Service string `json:"service"`
	OrigPkts *int `json:"orig_pkts"`

	// Http.log stuff
	//UserAgent string `json:"user_agent"`
	Method     *string `json:"method"`
	StatusCode *int    `json:"status_code"`

	// Dns.log stuff
	TransId *uint16 `json:"trans_id"`
	//RCode int `json:"rcode"`
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
