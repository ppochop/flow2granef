// package input provides readers of events from different technical types of input.
package input

import "github.com/prometheus/client_golang/prometheus"

type InputFactory func(InputConfig, InputStats) (Input, error)

type InputStats struct {
	MsgsConsumed prometheus.Counter
}

type Input interface {
	NextEntry() ([]byte, error)
}
type InputConfig map[string]any
