package structomancer

import (
	"reflect"
	"sync"
)

type (
	specCache struct {
		sync.RWMutex
		specs map[specCacheKey]*structSpec
	}

	specCacheKey struct {
		tagName    string
		structType reflect.Type
	}
)

var cache = newSpecCache()

func newSpecCache() *specCache {
	return &specCache{
		RWMutex: sync.RWMutex{},
		specs:   make(map[specCacheKey]*structSpec),
	}
}

func structSpecForType(t reflect.Type, tagName string) (spec *structSpec) {
	if !(IsStructType(t) || IsStructPtrType(t)) {
		panic("structomancer: unsupported type " + t.String())
	}

	key := specCacheKey{structType: t, tagName: tagName}

	cache.RLock()
	spec, found := cache.specs[key]
	cache.RUnlock()

	if found {
		return
	}

	cache.Lock()
	spec = newStructSpec(t, tagName)
	cache.specs[key] = spec
	cache.Unlock()

	return
}
