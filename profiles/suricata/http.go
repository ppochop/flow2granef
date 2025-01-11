package suricata

type SuricataHttp struct {
	Hostname  string `json:"hostname"`
	Url       string `json:"url"`
	UserAgent string `json:"http_user_agent"`
}
