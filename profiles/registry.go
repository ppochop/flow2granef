package profiles

import (
	"fmt"

	"github.com/dgraph-io/dgo/v240"
	xidcache "github.com/ppochop/flow2granef/xid-cache"
)

type transformersRegistryType map[string]TransformerFactory
type transformersDuplCheckRegistryType map[string]TransformerDuplCheckFactory

var transformersRegistry = make(transformersRegistryType)
var transformersDuplCheckRegistry = make(transformersDuplCheckRegistryType)

func GetTransformer(fSType string, cache *xidcache.IdCache, dgoClient *dgo.Dgraph, stats TransformerStats) (Transformer, error) {
	transformerFactory, found := transformersRegistry[fSType]
	if !found {
		return nil, fmt.Errorf("unknown profile %s", fSType)
	}
	return transformerFactory(cache, dgoClient, stats), nil
}

func GetTransformerDuplCheck(fSType string, cache *xidcache.DuplCache, instanceName string) (Transformer, error) {
	transformerFactory, found := transformersDuplCheckRegistry[fSType]
	if !found {
		return nil, fmt.Errorf("unknown profile %s", fSType)
	}
	return transformerFactory(cache, instanceName), nil
}

func RegisterTransformer(name string, tF TransformerFactory) {
	transformersRegistry[name] = tF
}

func RegisterDuplCheckTransformer(name string, tF TransformerDuplCheckFactory) {
	transformersDuplCheckRegistry[name] = tF
}
