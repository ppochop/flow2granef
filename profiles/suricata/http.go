package suricata

import "github.com/ppochop/flow2granef/flowutils"

type SuricataHttp struct {
	Hostname  string `json:"hostname"`
	Url       string `json:"url"`
	UserAgent string `json:"http_user_agent"`
}

func (s *SuricataHttp) GetGranefHTTPRec() *flowutils.HTTPRec {
	return &flowutils.HTTPRec{
		Hostname:  &s.Hostname,
		Url:       &s.Url,
		UserAgent: &s.UserAgent,
	}
}
