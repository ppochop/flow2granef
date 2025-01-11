package zeek

type ZeekLogType string

const (
	ZeekLogConn    ZeekLogType = "conn"
	ZeekLogDns     ZeekLogType = "dns"
	ZeekLogHttp    ZeekLogType = "http"
	ZeekLogUnknown ZeekLogType = "err"
)

type ZeekBase struct {
	// Conn.log stuff
	//Service string `json:"service"`
	OrigPkts *int `json:"orig_pkts"`

	// Http.log stuff
	//UserAgent string `json:"user_agent"`
	StatusCode *int `json:"status_code"`

	// Dns.log stuff
	QTypeName *string `json:"qtype_name"`
	//RCode int `json:"rcode"`
}

func (z *ZeekBase) decideType() ZeekLogType {
	switch {
	case z.OrigPkts != nil:
		return ZeekLogConn
	case z.StatusCode != nil:
		return ZeekLogHttp
	case z.QTypeName != nil:
		return ZeekLogDns
	default:
		return ZeekLogUnknown
	}
}
