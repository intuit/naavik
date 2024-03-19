package set

import "sync"

type Set[E comparable] interface {
	Add(item E)
	Clear()
	Delete(item E) bool
	Has(item E) bool
	Items() []E
	Size() int
}

type set[E comparable] struct {
	sync.RWMutex
	items map[E]struct{}
}

func NewSet[E comparable]() Set[E] {
	return &set[E]{
		items: map[E]struct{}{},
	}
}

func NewSetData[E comparable](data ...E) Set[E] {
	newset := &set[E]{
		items: map[E]struct{}{},
	}
	for _, item := range data {
		newset.Add(item)
	}
	return newset
}

func (s *set[E]) Add(item E) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.items[item]
	if !ok {
		s.items[item] = struct{}{}
	}
}

func (s *set[E]) Clear() {
	s.Lock()
	defer s.Unlock()
	s.items = make(map[E]struct{})
}

func (s *set[E]) Delete(item E) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.items[item]
	if ok {
		delete(s.items, item)
	}
	return ok
}

func (s *set[E]) Has(item E) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.items[item]
	return ok
}

func (s *set[E]) Items() []E {
	s.RLock()
	defer s.RUnlock()
	items := []E{}
	for i := range s.items {
		items = append(items, i)
	}
	return items
}

func (s *set[E]) Size() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.items)
}
