package main

import (
	"fmt"
)

type Set[T comparable] struct {
	M        map[T]interface{}
	SortFunc func(a, b T) int
}

func NewSet[T comparable](sortfunc func(a, b T) int) *Set[T] {
	return &Set[T]{
		M:        make(map[T]interface{}),
		SortFunc: sortfunc,
	}
}

func NewCoordSet() *Set[Coord] {
	return NewSet[Coord](func(a, b Coord) int {
		if a.Y < b.Y {
			return -1
		} else if a.Y > b.Y {
			return 1
		} else if a.X < b.X {
			return -1
		} else if a.X > b.X {
			return 1
		}
		return 0
	})
}

func NewNumSet(order int) *Set[int] {
	set := NewSet[int](func(a, b int) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	})
	for n := 1; n <= order; n++ {
		set.Add(n)
	}
	return set
}

func (s *Set[T]) String() string {
	out := ""
	idx := 0
	for t, _ := range s.M {
		if idx > 0 {
			out += " "
		}
		out += fmt.Sprintf("%s", t)
		idx++
	}
	return out
}

func (s *Set[T]) Size() int {
	return len(s.M)
}

func (s *Set[T]) GetOne() T {
	for k := range s.M {
		return k
	}
	panic(fmt.Errorf("trying to GetOne from empty set"))
}

func (s *Set[T]) Add(t T) bool {
	has := s.Has(t)
	s.M[t] = struct{}{}
	return !has
}

func (s *Set[T]) AddAll(o *Set[T]) bool {
	added := false
	for k := range o.M {
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

func (s *Set[T]) Clear() {
	for k, _ := range s.M {
		delete(s.M, k)
	}
}

func (s *Set[T]) Copy() *Set[T] {
	out := NewSet[T](s.SortFunc)
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

func (s *Set[T]) Equals(o *Set[T]) bool {
	if len(s.M) != len(o.M) {
		return false
	}
	for k := range s.M {
		if !o.Has(k) {
			return false
		}
	}
	return true
}

func (s *Set[T]) EqualsSlice(o []T) bool {
	if s.Size() != len(o) {
		return false
	}
	for _, v := range o {
		if !s.Has(v) {
			return false
		}
	}
	return true
}
