package suricata

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLoadEve(t *testing.T) {
	str := []byte(`{"timestamp":"2013-02-26T22:02:35.953172+0000","flow_id":507740360562687,"event_type":"flow","src_ip":"172.16.133.54","src_port":64347,"dest_ip":"173.194.75.104","dest_port":80,"proto":"TCP","flow":{"pkts_toserver":48,"pkts_toclient":58,"bytes_toserver":23192,"bytes_toclient":52436,"start":"2013-02-26T22:04:09.904649+0000","end":"2013-02-26T22:07:02.188702+0000","age":173,"state":"closed","reason":"shutdown","alerted":false},"tcp":{"tcp_flags":"1b","tcp_flags_ts":"1b","tcp_flags_tc":"1b","syn":true,"fin":true,"psh":true,"ack":true,"state":"closed","ts_max_regions":1,"tc_max_regions":1,"rst":null,"cwr":null},"app_proto":"http","icmp_type":null,"icmp_code":null,"response_icmp_type":null,"response_icmp_code":null,"app_proto_ts":null,"app_proto_tc":null}`)
	eve := SuricataEve{}
	err := json.Unmarshal(str, &eve)
	if err != nil {
		t.Fatalf("Failed to parse suricata json")
	}
	firstTs := time.Time{}
	lastTs := time.Time{}
	firstTs.UnmarshalText([]byte("2013-02-26T22:04:09.904649+00:00"))
	lastTs.UnmarshalText([]byte("2013-02-26T22:07:02.188702+00:00"))
	gotFirstTs := eve.Flow.GetFirstTs()
	gotLastTs := eve.Flow.GetLastTs()
	if !firstTs.Equal(gotFirstTs) {
		t.Fatalf("FirstTS not equal, expected %s, got %s", firstTs, gotFirstTs)
	}
	if !lastTs.Equal(gotLastTs) {
		t.Fatalf("LastTS not equal, expected %s, got %s", lastTs, gotLastTs)
	}
}
