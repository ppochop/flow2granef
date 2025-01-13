package ipfix

import (
	"encoding/json"
	"testing"

	"github.com/satta/gommunityid"
)

func TestCommId(t *testing.T) {
	f1 := []byte(`{"@type":"ipfix.entry","iana:flowEndReason":0,"iana:octetDeltaCount":64,"iana@reverse:octetDeltaCount@reverse":0,"iana:packetDeltaCount":1,"iana@reverse:packetDeltaCount@reverse":0,"iana:flowStartMicroseconds":"2013-02-26T22:02:56.271Z","iana:flowEndMicroseconds":"2013-02-26T22:02:56.271Z","iana:ipVersion":4,"iana:protocolIdentifier":"UDP","iana:tcpControlBits":"......","iana@reverse:tcpControlBits@reverse":"......","iana:sourceTransportPort":52853,"iana:destinationTransportPort":53,"iana:ingressInterface":0,"iana:sourceIPv4Address":"172.16.133.6","iana:destinationIPv4Address":"8.8.8.8","iana:sourceMacAddress":"00:19:B9:DA:15:A0","iana:destinationMacAddress":"00:90:7F:3E:02:D0","cesnet:DNSAnswers":0,"cesnet:DNSRCode":0,"cesnet:DNSQType":1,"cesnet:DNSClass":1,"cesnet:DNSRRTTL":0,"cesnet:DNSRDataLength":0,"cesnet:DNSPSize":0,"cesnet:DNSRDO":0,"cesnet:DNSTransactionID":51835,"cesnet:DNSName":"www.wip4.adobe.com","cesnet:DNSRData":"","iana:flowId":1046060674446557591}`)
	f2 := []byte(`{"@type":"ipfix.entry","iana:flowEndReason":0,"iana:octetDeltaCount":80,"iana@reverse:octetDeltaCount@reverse":0,"iana:packetDeltaCount":1,"iana@reverse:packetDeltaCount@reverse":0,"iana:flowStartMicroseconds":"2013-02-26T22:02:56.293Z","iana:flowEndMicroseconds":"2013-02-26T22:02:56.293Z","iana:ipVersion":4,"iana:protocolIdentifier":"UDP","iana:tcpControlBits":"......","iana@reverse:tcpControlBits@reverse":"......","iana:sourceTransportPort":53,"iana:destinationTransportPort":52853,"iana:ingressInterface":0,"iana:sourceIPv4Address":"8.8.8.8","iana:destinationIPv4Address":"172.16.133.6","iana:sourceMacAddress":"00:90:7F:3E:02:D0","iana:destinationMacAddress":"00:19:B9:DA:15:A0","cesnet:DNSAnswers":1,"cesnet:DNSRCode":0,"cesnet:DNSQType":1,"cesnet:DNSClass":1,"cesnet:DNSRRTTL":30,"cesnet:DNSRDataLength":13,"cesnet:DNSPSize":0,"cesnet:DNSRDO":0,"cesnet:DNSTransactionID":51835,"cesnet:DNSName":"www.wip4.adobe.com","cesnet:DNSRData":"192.150.16.64","iana:flowId":13220912849057427144}`)

	commIdGen, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	event1 := IpfixprobeFlow{}
	json.Unmarshal(f1, &event1)
	flow1 := event1.GetGranefFlowRec("ipfixprobe")
	ft1 := flow1.GetFlowTuple()

	commIdHash1 := commIdGen.Hash(ft1)
	commId1 := commIdGen.RenderBase64(commIdHash1)

	event2 := IpfixprobeFlow{}
	json.Unmarshal(f2, &event2)
	flow2 := event2.GetGranefFlowRec("ipfixprobe")
	ft2 := flow2.GetFlowTuple()

	commIdHash2 := commIdGen.Hash(ft2)
	commId2 := commIdGen.RenderBase64(commIdHash2)

	if commId1 != commId2 {
		t.Fatalf("inequal community ids for opposite directions of flow")
	}
}
