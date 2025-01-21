package zeek

import (
	"github.com/ppochop/flow2granef/flowutils"
)

type ZeekHttp struct {
	Host       string `json:"host"`
	Uri        string `json:"uri"`
	UserAgent  string `json:"user_agent"`
	Method     string `json:"method"`
	StatusCode uint16 `json:"status_code"`
}

func (z *ZeekHttp) GetGranefHTTPRec() *flowutils.HTTPRec {
	return &flowutils.HTTPRec{
		Hostname:   &z.Host,
		Url:        &z.Uri,
		UserAgent:  &z.UserAgent,
		Method:     &z.Method,
		StatusCode: z.StatusCode,
	}
}
