package main

type Set[T comparable] struct {
	M map[T]interface{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		M: make(map[T]interface{}),
	}
}

func NewCoordSet() *Set[Coord] {
	return NewSet[Coord]()
}

func (s *Set[T]) Size() int {
	return len(s.M)
}

func (s *Set[T]) Add(t T) bool {
	has := s.Has(t)
	s.M[t] = struct{}{}
	return !has
}

func (s *Set[T]) AddAll(o *Set[T]) bool {
	added := false
	for k, _ := range o.M {
		if s.Add(k) {
			added = true
		}
	}
	return added
}

func (s *Set[T]) Has(t T) bool {
	_, ok := s.M[t]
	return ok
}

func (s *Set[T]) Del(t T) bool {
	has := s.Has(t)
	delete(s.M, t)
	return has
}

func (s *Set[T]) Copy() *Set[T] {
	out := NewSet[T]()
	out.AddAll(s)
	return out
}

func (s *Set[T]) IntersectWith(o *Set[T]) bool {
	changed := false
	for k, _ := range s.M {
		if !o.Has(k) {
			s.Del(k)
			changed = true
		}
	}
	return changed
}
