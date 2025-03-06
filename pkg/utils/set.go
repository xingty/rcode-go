package utils

import "fmt"

const VALUE = true

type Set[T comparable] struct {
	m map[T]bool
}

func NewSet[T comparable](values ...T) *Set[T] {
	set := &Set[T]{m: make(map[T]bool)}
	for _, v := range values {
		set.m[v] = VALUE
	}
	return set
}

func (s *Set[T]) Add(value T) {
	s.m[value] = VALUE
}

func (s *Set[T]) Has(value T) bool {
	return s.m[value]
}

func (s *Set[T]) Remove(value T) {
	delete(s.m, value)
}

func (s *Set[T]) Size() int {
	return len(s.m)
}

func (s *Set[T]) Clear() {
	s.m = make(map[T]bool)
}

func (s *Set[T]) Values() []T {
	values := make([]T, 0, len(s.m))
	for value := range s.m {
		values = append(values, value)
	}
	return values
}

func (s *Set[T]) Intersect(other *Set[T]) *Set[T] {
	intersection := NewSet[T]()
	for value := range s.m {
		if other.Has(value) {
			intersection.Add(value)
		}
	}
	return intersection
}

func (s *Set[T]) String() string {
	return fmt.Sprintf("Set {%v}\n", s.Values())
}
