package utils

type XSet[T comparable] struct {
	values map[T]bool
}

func NewXSet[T comparable](elements ...T) *XSet[T] {
	set := &XSet[T]{
		values: make(map[T]bool),
	}
	for _, el := range elements {
		set.Add(el)
	}
	return set
}

func (s *XSet[T]) Add(element T) {
	s.values[element] = true
}

func (s *XSet[T]) Remove(element T) {
	delete(s.values, element)
}

func (s *XSet[T]) Contains(element T) bool {
	return s.values[element]
}

func (s *XSet[T]) Size() int {
	return len(s.values)
}

func (s *XSet[T]) IsEmpty() bool {
	return s.Size() > 0
}
