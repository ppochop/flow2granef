package zeek

import "encoding/json"

type ZeekHttp struct {
	Host      string `json:"host"`
	Uri       string `json:"uri"`
	UserAgent string `json:"user_agent"`
}

func (z *ZeekTransformer) ZeekHandleHttp(data []byte) error {
	connLimited := ZeekConnLimited{}
	http := ZeekHttp{}
	err := json.Unmarshal(data, &connLimited)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &http)
	if err != nil {
		return err
	}
	//send to granef
	return nil
}
