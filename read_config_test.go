package main_test

import (
	"testing"

	main "github.com/ppochop/flow2granef"
)

var data []byte = []byte(`
duplicity-check = false
passive-timeout = "10m"

[sources.suricata1]
transformer = "suricata"
input = "kafka"
workers-num = 4

[sources.suricata1.input-config]
bootstrap-servers = "localhost:9092"
group-id = "flow2granef-suricata1"
topic = "suricata1"

[sources.zeek1]
transformer = "zeek"
input = "kafka"
workers-num = 2

[sources.zeek1.input-config]
bootstrap-servers = "localhost:9092"
group-id = "flow2granef-zeek1"
topic = "zeek1"	
`)

func TestReadConfig(t *testing.T) {
	mC := main.MainConfig{}
	err := main.ReadConfig(data, &mC)
	if err != nil {
		t.Fatalf("Failed to read config.")
	}

}
