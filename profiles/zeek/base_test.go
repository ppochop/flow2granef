package zeek

import (
	"encoding/json"
	"testing"
)

func TestDecideTypeConn(t *testing.T) {
	zB := ZeekBase{}
	connStr := []byte(`{"ts":1361916250.722537,"uid":"CaqgMn4RR46OIO3Kzk","id.orig_h":"172.16.133.54","id.orig_p":64367,"id.resp_h":"54.235.184.94","id.resp_p":80,"proto":"tcp","service":"http","duration":59.16254496574402,"orig_bytes":467,"resp_bytes":465,"conn_state":"SF","local_orig":true,"local_resp":false,"missed_bytes":0,"history":"ShADadtfF","orig_pkts":6,"orig_ip_bytes":731,"resp_pkts":6,"resp_ip_bytes":1182}`)
	json.Unmarshal(connStr, &zB)
	res := zB.decideType()
	if res != ZeekLogConn {
		t.Fatalf("DecideType test failed, wanted %+v, got %+v", ZeekLogConn, res)
	}
}

func TestDecideTypeDns(t *testing.T) {
	zB := ZeekBase{}
	dnsStr := []byte(`{"ts":1361916329.622777,"uid":"CS814Y1mWvj72ZKKKe","id.orig_h":"172.16.133.41","id.orig_p":61319,"id.resp_h":"172.16.128.202","id.resp_p":53,"proto":"udp","trans_id":10425,"rtt":0.0955190658569336,"query":"www.travelocity.com","qclass":1,"qclass_name":"C_INTERNET","qtype":1,"qtype_name":"A","rcode":0,"rcode_name":"NOERROR","AA":false,"TC":false,"RD":true,"RA":true,"Z":0,"answers":["199.204.31.83"],"TTLs":[26.0],"rejected":false}`)
	json.Unmarshal(dnsStr, &zB)
	res := zB.decideType()
	if res != ZeekLogDns {
		t.Fatalf("DecideType test failed, wanted %+v, got %+v", ZeekLogDns, res)
	}
}

func TestDecideTypeHttp(t *testing.T) {
	zB := ZeekBase{}
	httpStr := []byte(`{"ts":1361916250.772851,"uid":"CVj62r2fPvBBQf8K4h","id.orig_h":"172.16.133.54","id.orig_p":64368,"id.resp_h":"173.204.219.5","id.resp_p":80,"trans_depth":1,"method":"GET","host":"p.brilig.com","uri":"/contact/bct?pid=a23194cb-a3d3-4c83-a6db-38b34fa79bbb&_ct=pixel&puid=0qm5V73&preto=&sedu=&pdegr=&sjgrp=&sind=&pind=","referrer":"http://www.salary.com/","version":"1.1","user_agent":"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:19.0) Gecko/20100101 Firefox/19.0","request_body_len":0,"response_body_len":43,"status_code":200,"status_msg":"OK","tags":[],"resp_fuids":["FtWdLQ1fbQqHFPJQ7b"],"resp_mime_types":["image/gif"]}`)
	json.Unmarshal(httpStr, &zB)
	res := zB.decideType()
	if res != ZeekLogHttp {
		t.Fatalf("DecideType test failed, wanted %+v, got %+v", ZeekLogHttp, res)
	}
}

func TestDecideTypeUnknown(t *testing.T) {
	zB := ZeekBase{}
	unStr := []byte(`{"ts":1361916253.362444,"uid":"CTvHK71LsMV4c26tTf","id.orig_h":"172.16.133.120","id.orig_p":53453,"id.resp_h":"96.43.146.22","id.resp_p":443,"version":"TLSv10","cipher":"TLS_RSA_WITH_RC4_128_MD5","server_name":"0.umps2c2.salesforce.com","resumed":true,"established":true,"ssl_history":"CsiI"}`)
	json.Unmarshal(unStr, &zB)
	res := zB.decideType()
	if res != ZeekLogUnknown {
		t.Fatalf("DecideType test failed, wanted %s, got %s", ZeekLogUnknown, res)
	}
}
