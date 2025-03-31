package input

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

type FileInput struct {
	fileScanner  bufio.Scanner
	read         uint
	msgsConsumed prometheus.Counter
}

func init() {
	RegisterInput("file", InitFileInput)
}

type FileConfig struct {
	Path string `toml:"path"`
}

func InitFileInput(config InputConfig, stats InputStats) (Input, error) {
	fC := FileConfig{
		Path: config["path"].(string),
	}
	file, err := os.Open(fC.Path)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s", fC.Path)
	}
	return &FileInput{
		fileScanner:  *bufio.NewScanner(file),
		msgsConsumed: stats.MsgsConsumed,
	}, nil
}

func (s *FileInput) NextEntry() ([]byte, error) {
	if s.fileScanner.Scan() {
		s.msgsConsumed.Inc()
		return []byte(s.fileScanner.Text()), nil
	}
	err := s.fileScanner.Err()
	if err != nil {
		slog.Error("Error when reading input", "error", err)
	} else {
		slog.Info("EOF")
		err = fmt.Errorf("EOF")
	}
	return nil, err
}

func (s *FileInput) GetStats() map[string]uint {
	return map[string]uint{
		"messages_read": s.read,
	}
}
