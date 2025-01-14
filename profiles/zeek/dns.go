package zeek

import (
	"net/netip"

	"github.com/ppochop/flow2granef/flowutils"
)

type ZeekDns struct {
	TransId uint16    `json:"trans_id"`
	Query   *string   `json:"query"`
	Answers []string  `json:"answers"`
	TTLs    []float64 `json:"ttls"`
	QType   *string   `json:"qtype_name"`
	RCode   *string   `json:"rcode_name"`
}

func (z *ZeekDns) GetGranefDNSRec() *flowutils.DNSRec {
	ret := &flowutils.DNSRec{
		TransId: &z.TransId,
		Query:   z.Query,
	}
	for i, ansStr := range z.Answers {
		ip, err := netip.ParseAddr(ansStr)
		if err != nil {
			continue
		}
		ret.Answer = append(ret.Answer, &ip)
		TTLuint := uint(z.TTLs[i])
		ret.TTL = append(ret.TTL, &TTLuint)
	}
	return ret
}
