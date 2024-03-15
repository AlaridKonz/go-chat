package datastructures

type XSet[T comparable] struct {
	values map[T]bool
}

func (s *XSet[T]) Init() *XSet[T] {
	s.values = make(map[T]bool)
	return s
}

func NewXSet[T comparable](elements ...T) *XSet[T] {
	set := new(XSet[T]).Init()
	set.AddAll(elements...)
	return set
}

func (s *XSet[T]) AddAll(elements ...T) {
	for _, el := range elements {
		s.Add(el)
	}
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

func (s *XSet[T]) ForEach(consumer func(T)) {
	for key, _ := range s.values {
		consumer(key)
	}
}
