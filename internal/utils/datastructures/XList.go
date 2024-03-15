package datastructures

type XElement[T any] struct {
	prev, next *XElement[T]

	list *XList[T]

	Value T
}

func (el *XElement[T]) Next() *XElement[T] {
	if next := el.next; el.list != nil && next != &el.list.root {
		return next
	}
	return nil
}

func (el *XElement[T]) Prev() *XElement[T] {
	if prev := el.prev; el.list != nil && prev != &el.list.root {
		return prev
	}
	return nil
}

type XList[T any] struct {
	root XElement[T]
	size int
}

func (l *XList[T]) Init() *XList[T] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.size = 0
	return l
}

func NewXList[T any]() *XList[T] {
	return new(XList[T]).Init()
}

func (l *XList[T]) Front() *XElement[T] {
	if l.size == 0 {
		return nil
	}
	return l.root.next
}

func (l *XList[T]) Back() *XElement[T] {
	if l.size == 0 {
		return nil
	}
	return l.root.prev
}

func (l *XList[T]) Size() int {
	return l.size
}

func (l *XList[T]) Remove(element *XElement[T]) *XElement[T] {
	if element.list == l {
		l.remove(element)
	}
	return nil
}

func (l *XList[T]) RemoveIf(test func(T) bool) int {
	elementsRemoved := 0
	for e := l.Front(); e != nil; e = e.Next() {
		if test(e.Value) {
			l.Remove(e)
			elementsRemoved++
		}
	}
	return elementsRemoved
}

func (l *XList[T]) RemoveFirst(value T) *XElement[T] {
	for e := l.Front(); e != nil; e = e.Next() {
		if &value == &e.Value {
			return l.Remove(e)
		}
	}
	return nil
}

func (l *XList[T]) RemoveLast(value T) *XElement[T] {
	for e := l.Back(); e != nil; e = e.Prev() {
		if &value == &e.Value {
			return l.Remove(e)
		}
	}
	return nil
}

func (l *XList[T]) Contains(value T) bool {
	containsIt := false
	l.ForEach(func(el T) {
		if &el == &value {
			containsIt = true
			return
		}
	})
	return containsIt
}

func (l *XList[T]) PushFront(value T) *XElement[T] {
	l.backupInit()
	return l.insertValue(value, &l.root)
}

func (l *XList[T]) PushBack(value T) *XElement[T] {
	l.backupInit()
	return l.insertValue(value, l.root.prev)

}

func (l *XList[T]) PushFrontList(other *XList[T]) {
	l.backupInit()
	for c, el := other.size, other.Front(); c > 0; c, el = c-1, el.Next() {
		l.insertValue(el.Value, &l.root)
	}
}

func (l *XList[T]) PushBackList(other *XList[T]) {
	l.backupInit()
	for c, el := other.size, other.Back(); c > 0; c, el = c-1, el.Prev() {
		l.insertValue(el.Value, l.root.prev)
	}
}

func (l *XList[T]) InsertBefore(v T, mark *XElement[T]) *XElement[T] {
	if mark.list != l {
		return nil
	}
	return l.insertValue(v, mark.prev)
}

func (l *XList[T]) InsertAfter(v T, mark *XElement[T]) *XElement[T] {
	if mark.list != l {
		return nil
	}
	return l.insertValue(v, mark)
}

func (l *XList[T]) MoveToFront(e *XElement[T]) {
	if e.list != l || l.root.next == e {
		return
	}
	l.move(e, &l.root)
}

func (l *XList[T]) MoveToBack(e *XElement[T]) {
	if e.list != l || l.root.prev == e {
		return
	}
	l.move(e, l.root.prev)
}

func (l *XList[T]) MoveBefore(e, mark *XElement[T]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

func (l *XList[T]) MoveAfter(e, mark *XElement[T]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

func (l *XList[T]) move(e, at *XElement[T]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

func (l *XList[T]) insertValue(val T, pos *XElement[T]) *XElement[T] {
	return l.insertAt(&XElement[T]{Value: val}, pos)
}

func (l *XList[T]) insertAt(el, pos *XElement[T]) *XElement[T] {
	el.next = pos.next
	el.prev = pos
	el.next.prev = el
	el.prev.next = el
	el.list = l
	l.size++
	return el
}

func (l *XList[T]) backupInit() {
	if l.root.next == nil {
		l.Init()
	}
}

func (l *XList[T]) remove(element *XElement[T]) {
	element.prev.next = element.next
	element.next.prev = element.prev
	element.prev = nil
	element.next = nil
	element.list = nil
	l.size--
}

func (l *XList[T]) ForEach(consumer func(T)) {
	for c, el := l.Size(), l.Front(); c > 0; c, el = c-1, el.Next() {
		consumer(el.Value)
	}
}

func (l *XList[T]) ForEachReversed(consumer func(T)) {
	for c, el := l.Size(), l.Back(); c > 0; c, el = c-1, el.Prev() {
		consumer(el.Value)
	}
}

func (l *XList[T]) IsEmpty() bool {
	return l.Size() == 0
}

func MapList[T, U any](l *XList[T], transform func(T) U) *XList[U] {
	newList := new(XList[U]).Init()
	for c, el := l.Size(), l.Front(); c > 0; c, el = c-1, el.Next() {
		newValue := transform(el.Value)
		newList.PushFront(newValue)
	}
	return newList
}
