package main

import (
)

type Iter[T any] interface {
	Iterate(yield func(T) error) error
}

type Iterator[T any] struct {
	Iteratef func(yield func(T) error) error
	Closef func() error
}

func (i *Iterator[T]) Iterate(yield func(T) error) error {
	return i.Iteratef(yield)
}

func (i *Iterator[T]) Close() error {
	if i.Closef != nil {
		return i.Closef()
	}
	return nil
}

type Puller[T any] interface {
	Next() (T, error)
}

type PushPull[T any] struct {
	Nextf func() (T, error)
	Closef func() error
}

func (p *PushPull[T]) Next() (T ,error) {
	return p.Nextf()
}

func (p *PushPull[T]) Close() error {
	return p.Closef()
}
