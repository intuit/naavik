package cache

import (
	"context"
	"strings"
	"sync"
)

type ControllerCacheInterface interface {
	BaseCache
	Register(name string, ctlrContext ControllerContext)
	DeRegister(name string)
	List() map[string]ControllerContext
	GetStopCh(name string) ControllerContext
	Range(f func(key any, value interface{}) bool)
}

type ctlrCache struct {
	name string
}

type ControllerContext struct {
	WorkerCtx []context.Context
	StopCh    chan struct{}
}

var ControllerCache = newControllerCache()

var controllerCacheMap = sync.Map{}

func newControllerCache() ControllerCacheInterface {
	return ctlrCache{}
}

func (ctlrCache) Register(name string, ctlrContext ControllerContext) {
	name = strings.ToLower(name)
	controllerCacheMap.Store(name, ctlrContext)
}

func (ctlrCache) DeRegister(name string) {
	name = strings.ToLower(name)
	controllerCacheMap.Delete(name)
}

func (ctlrCache) List() map[string]ControllerContext {
	ctlrMap := make(map[string]ControllerContext)
	controllerCacheMap.Range(func(key, value interface{}) bool {
		ctlrMap[key.(string)] = value.(ControllerContext)
		return true
	})
	return ctlrMap
}

func (ctlrCache) GetStopCh(name string) ControllerContext {
	name = strings.ToLower(name)
	stopCh, ok := controllerCacheMap.Load(name)
	if !ok {
		return ControllerContext{}
	}
	return stopCh.(ControllerContext)
}

func (ctlrCache) Range(f func(key, value interface{}) bool) {
	controllerCacheMap.Range(f)
}

func (ctlrCache) Reset() {
	controllerCacheMap = sync.Map{}
}
