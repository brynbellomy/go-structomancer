package structomancer

import (
	"reflect"
	"sync"
)

type (
	SpecCache struct {
		sync.RWMutex
		specs map[SpecCacheKey]*StructSpec
	}

	SpecCacheKey struct {
		tagName    string
		structType reflect.Type
	}
)

var specCache = newSpecCache()

func newSpecCache() *SpecCache {
	return &SpecCache{
		RWMutex: sync.RWMutex{},
		specs:   make(map[SpecCacheKey]*StructSpec),
	}
}

func StructSpecForType(specimen interface{}, tagName string) (spec *StructSpec) {
	t := EnsureStructOrStructPointerType(reflect.TypeOf(specimen))

	key := SpecCacheKey{structType: t, tagName: tagName}

	specCache.RLock()
	spec, found := specCache.specs[key]
	specCache.RUnlock()

	if found {
		return
	}

	specCache.Lock()
	spec = NewStructSpec(t, tagName)
	specCache.specs[key] = spec
	specCache.Unlock()

	return
}
