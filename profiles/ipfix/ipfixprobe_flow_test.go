package ipfix

import (
	"encoding/json"
	"testing"
)

func TestFlowParse(t *testing.T) {
	flowStr := []byte(`{"@type":"ipfix.entry","iana:flowEndReason":0,"iana:octetDeltaCount":71,"iana@reverse:octetDeltaCount@reverse":0,"iana:packetDeltaCount":1,"iana@reverse:packetDeltaCount@reverse":0,"iana:flowStartMicroseconds":"2013-02-26T22:02:36.058Z","iana:flowEndMicroseconds":"2013-02-26T22:02:36.058Z","iana:ipVersion":4,"iana:protocolIdentifier":"UDP","iana:tcpControlBits":"......","iana@reverse:tcpControlBits@reverse":"......","iana:sourceTransportPort":63229,"iana:destinationTransportPort":53,"iana:ingressInterface":0,"iana:sourceIPv4Address":"172.16.133.6","iana:destinationIPv4Address":"8.8.8.8","iana:sourceMacAddress":"00:19:B9:DA:15:A0","iana:destinationMacAddress":"00:90:7F:3E:02:D0","cesnet:DNSAnswers":0,"cesnet:DNSRCode":0,"cesnet:DNSQType":12,"cesnet:DNSClass":1,"cesnet:DNSRRTTL":0,"cesnet:DNSRDataLength":0,"cesnet:DNSPSize":0,"cesnet:DNSRDO":0,"cesnet:DNSTransactionID":36336,"cesnet:DNSName":"45.66.120.96.in-addr.arpa","cesnet:DNSRData":"","iana:flowId":3113537407031653662}`)
	flow := IpfixprobeFlow{}
	err := json.Unmarshal(flowStr, &flow)
	if err != nil {
		t.Fatalf("Failed to parse flow")
	}
	isDnsAnswer := flow.IsDnsAnswer()
	if isDnsAnswer {
		t.Fatalf("Falsely determined to be DNS answer")
	}
}
