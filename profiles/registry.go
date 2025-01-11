package profiles

import (
	"fmt"

	"github.com/dgraph-io/dgo/v240"
)

type transformersRegistryType map[string]TransformerFactory
type transformersDuplCheckRegistryType map[string]TransformerDuplCheckFactory
type preHandlerRegistryType map[string]PreHandler

var transformersRegistry = make(transformersRegistryType)
var transformersDuplCheckRegistry = make(transformersDuplCheckRegistryType)
var preHandlerRegistry = make(preHandlerRegistryType)

func GetTransformer(fSType string, cache Cache, dgoClient *dgo.Dgraph, stats TransformerStats) (Transformer, error) {
	transformerFactory, found := transformersRegistry[fSType]
	if !found {
		return nil, fmt.Errorf("unknown profile %s", fSType)
	}
	return transformerFactory(cache, dgoClient, stats), nil
}

func GetTransformerDuplCheck(fSType string, cache CacheDuplCheck, instanceName string) (Transformer, error) {
	transformerFactory, found := transformersDuplCheckRegistry[fSType]
	if !found {
		return nil, fmt.Errorf("unknown profile %s", fSType)
	}
	return transformerFactory(cache, instanceName), nil
}

func GetPreHandler(fSType string) (PreHandler, error) {
	preHandler, found := preHandlerRegistry[fSType]
	if !found {
		return nil, fmt.Errorf("unknown profile %s", fSType)
	}
	return preHandler, nil
}

func RegisterTransformer(name string, tF TransformerFactory) {
	transformersRegistry[name] = tF
}

func RegisterDuplCheckTransformer(name string, tF TransformerDuplCheckFactory) {
	transformersDuplCheckRegistry[name] = tF
}

func RegisterPreHandler(name string, p PreHandler) {
	preHandlerRegistry[name] = p
}
