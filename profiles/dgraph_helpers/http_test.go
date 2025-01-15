package dgraphhelpers

import "testing"

func TestHttpPaths(t *testing.T) {
	url1 := `/scds/concat/common/js?h=em1445ff7g1bpi7y8m5y0fje5-bw30xicxn7t2ahoe5bs20g38b-6olwp79c7gbdw66zec9vm1ave-aaod7sf2exa7qhhbzubedjht1-awyr8kgquw0r9wjm5o3ness8g-9fn1u8cisjms2qtlsya7r23k7-8v6o0480wy5u6j7f3sh92hzxo-dyt8o4nwtaujeutlgncuqe0dn-dbfbjb67f2aszyqrhgq65b9c6-887xyt8ztt77epd54gc10fdzy-4evljp80k0uxyvh5vslbdawkk-3tu63s1gnqwg9wt4x4xxbmnat-624brk691lqhhqtdw3ai6lss6-7doxw9g7rbv4rtn9dq396yn9t-ai9mq0laqs728dkzr84zl6slm-d7pcuzjq8dg78xeof54dglz9v-2ri3p7q5yuaxtesgtaa4d5x44-8dfmf7pw4cuq5uwk1ndzx3jdf-1up880t5llbsqhsngxq05ez9i-bsnr81j3v60h85h0kpvj4ednu-aykdde9kbba63phksx4ysarff-496kwpvz3hjd061ql0urz3j1z-dq10seyrd8v4vegmjyfk6hl5r-4zd16izxsd0kpjsu30jhu2nco-aqcdjoyfx9mrkza4xqv8qnc4f-1ifz6mbjq4m4m0tqy526sdot-6o3io0gmt64mljaiicaw4m51p&fc=2`

	url2 := `/scds/concat/common/js`
	_, path := handleUrl(&url1)
	if *path != "/scds/concat/common/js" {
		t.Fatalf("inequal paths")
	}

	_, path = handleUrl(&url2)
	if *path != "/scds/concat/common/js" {
		t.Fatalf("inequal paths")
	}
}
