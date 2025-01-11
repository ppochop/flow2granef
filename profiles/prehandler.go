package profiles

type PreHandler func(flow []byte) (uint32, error)
