package input

import (
	"fmt"
)

type inputRegistryType map[string]InputFactory

var inputRegistry = make(inputRegistryType)

func GetInput(iType string, config InputConfig, stats InputStats) (Input, error) {
	inputFactory, found := inputRegistry[iType]
	if !found {
		return nil, fmt.Errorf("unknown inputter %s", iType)
	}
	return inputFactory(config, stats)
}

func RegisterInput(name string, iF InputFactory) {
	inputRegistry[name] = iF
}
