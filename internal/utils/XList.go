package utils

import "container/list"

type XList struct {
	*list.List
}

func (l *XList) RemoveIf(test func(value interface{}) bool) any {
	for e := l.Front(); e.Value != nil; e = e.Next() {
		if test(e) {
			return l.Remove(e)
		}
	}
	return nil
}

func (l *XList) RemoveFirst(value interface{}) any {
	for e := l.Front(); e.Value != nil; e = e.Next() {
		if value == e.Value {
			return l.Remove(e)
		}
	}
	return nil
}

func (l *XList) RemoveLast(value interface{}) any {
	for e := l.Back(); e.Value != nil; e = e.Prev() {
		if value == e.Value {
			return l.Remove(e)
		}
	}
	return nil
}
