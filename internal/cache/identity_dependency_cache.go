package cache

import (
	"strings"
	"sync"

	"github.com/intuit/naavik/pkg/types/set"
)

type IdentityDependencyCache interface {
	BaseCache
	// AddDependencyToIdentity adds a dependency to the given identity
	AddDependencyToIdentity(identity string, dependency string)
	// AddDependentToIdentity adds a dependent to the given identity
	AddDependentToIdentity(identity string, dependent string)
	DeleteDependencyFromIdentity(identity string, dependency string)
	DeleteDependentFromIdentity(identity string, dependent string)
	// GetDependenciesForIdentity returns a list of identities that the given identity depends on
	GetDependenciesForIdentity(identity string) []string
	// GetDependentsForIdentity returns a list of identities that are dependent on the given identity
	GetDependentsForIdentity(identity string) []string
	IsDependentForIdentity(identity string, dependency string) bool
	GetTotalDependencies() int
	RangedDependencies(func(identity string, dependencies []string) bool)

	// Used for unit tests, not advised to use this in production code
	Reset()
}

var IdentityDependency = newIdentityDependency()

type identityDependency struct{}

var (
	identityDependencyCacheLock = sync.RWMutex{}
	identityDependentCacheLock  = sync.RWMutex{}
	identityDependencyCache     = sync.Map{} // Map[identity]Set[dependency]
	identityDependentCache      = sync.Map{} // Map[identity]Set[dependent]
)

func newIdentityDependency() IdentityDependencyCache {
	return &identityDependency{}
}

func (*identityDependency) AddDependencyToIdentity(identity string, dependency string) {
	identity = strings.ToLower(identity)
	dependency = strings.ToLower(dependency)
	identityDependencyCacheLock.Lock()
	defer identityDependencyCacheLock.Unlock()
	existing, loaded := identityDependencyCache.Load(identity)
	if loaded {
		existing.(set.Set[string]).Add(dependency)
		identityDependencyCache.Store(identity, existing)
	} else {
		dependencies := set.NewSet[string]()
		dependencies.Add(dependency)
		identityDependencyCache.Store(identity, dependencies)
	}
}

func (*identityDependency) AddDependentToIdentity(identity string, dependent string) {
	identity = strings.ToLower(identity)
	dependent = strings.ToLower(dependent)
	identityDependentCacheLock.Lock()
	defer identityDependentCacheLock.Unlock()
	existing, loaded := identityDependentCache.Load(identity)
	if loaded {
		existing.(set.Set[string]).Add(dependent)
		identityDependentCache.Store(identity, existing)
	} else {
		dependents := set.NewSet[string]()
		dependents.Add(dependent)
		identityDependentCache.Store(identity, dependents)
	}
}

func (*identityDependency) DeleteDependencyFromIdentity(identity string, dependency string) {
	identity = strings.ToLower(identity)
	dependency = strings.ToLower(dependency)
	identityDependencyCacheLock.Lock()
	defer identityDependencyCacheLock.Unlock()
	existing, found := identityDependencyCache.Load(identity)
	if !found {
		return
	}
	dependencies := existing.(set.Set[string])
	dependencies.Delete(dependency)
	if dependencies.Size() == 0 {
		identityDependencyCache.Delete(identity)
		return
	}
	identityDependencyCache.Store(identity, dependencies)
}

func (*identityDependency) DeleteDependentFromIdentity(identity string, dependent string) {
	identity = strings.ToLower(identity)
	dependent = strings.ToLower(dependent)
	identityDependentCacheLock.Lock()
	defer identityDependentCacheLock.Unlock()
	existing, found := identityDependentCache.Load(identity)
	if !found {
		return
	}
	dependents := existing.(set.Set[string])
	dependents.Delete(dependent)
	if dependents.Size() == 0 {
		identityDependentCache.Delete(identity)
		return
	}
	identityDependentCache.Store(identity, dependents)
}

func (identityDependency) GetDependenciesForIdentity(identity string) []string {
	identity = strings.ToLower(identity)
	identityDependencyCacheLock.Lock()
	defer identityDependencyCacheLock.Unlock()
	existing, found := identityDependencyCache.Load(identity)
	if !found {
		return []string{}
	}
	existingDependencies := existing.(set.Set[string]).Items()
	clone := make([]string, len(existingDependencies))
	copy(clone, existingDependencies)
	return clone
}

func (identityDependency) GetDependentsForIdentity(identity string) []string {
	identity = strings.ToLower(identity)
	identityDependentCacheLock.RLock()
	defer identityDependentCacheLock.RUnlock()
	existing, found := identityDependentCache.Load(identity)
	if !found {
		return []string{}
	}
	existingDependents := existing.(set.Set[string]).Items()
	clone := make([]string, len(existingDependents))
	copy(clone, existingDependents)
	return clone
}

func (ic identityDependency) IsDependentForIdentity(identity string, dependency string) bool {
	identity = strings.ToLower(identity)
	dependency = strings.ToLower(dependency)
	identityDependencyCacheLock.RLock()
	defer identityDependencyCacheLock.RUnlock()
	existing, found := identityDependencyCache.Load(identity)
	if !found {
		return false
	}
	return existing.(set.Set[string]).Has(dependency)
}

func (identityDependency) GetTotalDependencies() int {
	total := 0
	identityDependencyCache.Range(func(key, value interface{}) bool {
		total++
		return true
	})
	return total
}

func (identityDependency) RangedDependencies(f func(identity string, dependencies []string) bool) {
	identityDependencyCache.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(set.Set[string]).Items())
	})
}

func (identityDependency) Reset() {
	identityDependencyCacheLock.Lock()
	defer identityDependencyCacheLock.Unlock()
	identityDependentCacheLock.Lock()
	defer identityDependentCacheLock.Unlock()
	identityDependencyCache = sync.Map{}
	identityDependentCache = sync.Map{}
}
