package zeek

import "encoding/json"

type ZeekDns struct {
	TransId uint16   `json:"trans_id"`
	Query   *string  `json:"query"`
	Answers []string `json:"answers"`
	TTLs    []int    `json:"ttls"`
	QType   *string  `json:"qtype_name"`
	RCode   *string  `json:"rcode_name"`
}

func (z *ZeekTransformer) ZeekHandleDns(data []byte) error {
	connLimited := ZeekConnLimited{}
	dns := ZeekDns{}
	err := json.Unmarshal(data, &connLimited)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &dns)
	if err != nil {
		return err
	}
	//send to granef
	return nil
}
