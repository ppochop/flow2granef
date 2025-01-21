package suricata

import (
	"net/netip"

	"github.com/ppochop/flow2granef/flowutils"
)

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
	Version uint   `json:"version"`
	Type    string `json:"type"`
	Id      uint16 `json:"id"`
	RCode   string `json:"rcode"`
	RRName  string `json:"rrname"`
	RRType  string `json:"rrtype"`
	//Queries []SuricataDnsQuery  `json:"queries"`
	Answers []SuricataDnsAnswer `json:"answers"`
}

func (s *SuricataDnsInfo) GetGranefDNSRec() *flowutils.DNSRec {
	ret := &flowutils.DNSRec{
		TransId: &s.Id,
		Query:   &s.RRName,
		QType:   &s.RRType,
	}
	for _, ans := range s.Answers {
		if ans.RRtype != "A" && ans.RRtype != "AAAA" {
			continue
		}
		ip, err := netip.ParseAddr(ans.RData)
		if err != nil {
			continue
		}
		ret.Answer = append(ret.Answer, &ip)
		ret.TTL = append(ret.TTL, &ans.Ttl)
	}
	return ret
}
