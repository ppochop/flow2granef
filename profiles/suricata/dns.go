package suricata

type SuricataDnsQuery struct {
	RRName string `json:"rrname"`
	RRType string `json:"rrtype"`
}

type SuricataDnsAnswer struct {
	RRName string `json:"rrname"`
	RRtype string `json:"rrtype"`
	Ttl    uint   `json:"ttl"`
	RData  string `json:"rdata"`
}

type SuricataDnsInfo struct {
	Version uint                `json:"version"`
	Type    string              `json:"type"`
	Id      uint                `json:"id"`
	RCode   string              `json:"rcode"`
	Queries []SuricataDnsQuery  `json:"queries"`
	Answers []SuricataDnsAnswer `json:"answers"`
}
