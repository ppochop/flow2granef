package flowutils

type FlushReason string

const (
	ActiveTimeout  FlushReason = "a"
	PassiveTimeout FlushReason = "p"
	Finished       FlushReason = "f"
	Unknown        FlushReason = "?"
)
