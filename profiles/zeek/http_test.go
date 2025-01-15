package zeek

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestHttp(t *testing.T) {
	hEv := []byte(`{"ts":1736848105.147166,"uid":"CPJ1fj2NeOkljYAAIf","id.orig_h":"172.16.133.132","id.orig_p":52333,"id.resp_h":"98.139.240.23","id.resp_p":80,"trans_depth":1,"method":"GET","host":"us.bc.yahoo.com","uri":"/b?P=CWSMYDc2LjH239XVUS0w1RBTNTAuN1EtMP3__70k&T=18088ib1p/X=1361916157/E=76001284/R=network/K=5/V=8.1/W=0/Y=yahoo/F=4203755486/H=YWRjdmVyPSI2LjQuNCIgc2VydmVJZD0iQ1dTTVlEYzJMakgyMzlYVlVTMHcxUkJUTlRBdU4xRXRNUDNfXzcwayIgdFN0bXA9IjEzNjE5MTYxNTc3NzM3MDEiIA--/Q=-1/S=1/J=69060D4C&U=12b356idu/N=TvffBUwNPRU-/C=-2/D=MON/B=-2/V=0","referrer":"http://webhosting.yahoo.com/forward.html","version":"1.0","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/534.55.3 (KHTML, like Gecko) Version/5.1.3 Safari/534.53.10","request_body_len":0,"response_body_len":43,"status_code":200,"status_msg":"OK","tags":[],"resp_fuids":["FYLfJc36OQO3ud6xZd"],"resp_mime_types":["image/gif"]}`)
	http := ZeekHttp{}
	json.Unmarshal(hEv, &http)
	gHttp := http.GetGranefHTTPRec()
	resp := fmt.Sprintf(`asdf "%s"`, *gHttp.Url)
	if *gHttp.Hostname != "us.bc.yahoo.com" {
		t.Fatalf("bad hostname %s", resp)
	}
}
