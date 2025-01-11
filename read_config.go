package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type SourceConfig struct {
	WorkersNum      uint           `toml:"workers-num"`
	TransformerName string         `toml:"transformer"`
	InputName       string         `toml:"input"`
	InputConfig     map[string]any `toml:"input-config"`
}

type MainConfig struct {
	DuplCheck      bool                    `toml:"duplicity-check"`
	PassiveTimeout string                  `toml:"passive-timeout"`
	DgraphAddress  string                  `toml:"dgraph-address"`
	ResetDgraph    bool                    `toml:"reset-dgraph"`
	Sources        map[string]SourceConfig `toml:"sources"`
}

func ReadConfigFile(path string) (MainConfig, error) {
	mC := MainConfig{}
	config, err := os.ReadFile(path)
	if err != nil {
		return mC, fmt.Errorf("failed to open file %s", path)
	}
	return mC, ReadConfig(config, &mC)
}

func ReadConfig(file []byte, mC *MainConfig) error {
	err := toml.Unmarshal(file, mC)
	if err != nil {
		return fmt.Errorf("failed to parse config")
	}
	return nil
}
