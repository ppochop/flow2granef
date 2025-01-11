package input

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

type StdinInput struct {
	stdin        bufio.Scanner
	read         uint
	msgsConsumed prometheus.Counter
}

func init() {
	RegisterInput("stdin", InitStdinInput)
}

func InitStdinInput(config InputConfig, stats InputStats) (Input, error) {
	return &StdinInput{
		stdin:        *bufio.NewScanner(os.Stdin),
		msgsConsumed: stats.MsgsConsumed,
	}, nil
}

func (s *StdinInput) NextEntry() ([]byte, error) {
	if s.stdin.Scan() {
		s.msgsConsumed.Inc()
		return s.stdin.Bytes(), nil
	}
	err := s.stdin.Err()
	if err != nil {
		slog.Error("Error when reading input", "error", err)
	} else {
		slog.Info("EOF")
		err = fmt.Errorf("EOF")
	}
	return nil, err
}

func (s *StdinInput) GetStats() map[string]uint {
	return map[string]uint{
		"messages_read": s.read,
	}
}
