package main

import (
	"io"
	"sync"
	"errors"
)

func CloseAny(as ...any) error {
	var err error
	for _, a := range as {
		c, ok := a.(io.Closer)
		if ok {
			e := c.Close()
			if err == nil {
				err = e
			}
		}
	}
	return err
}

func Push[T any](p Puller[T]) *Iterator[T] {
	iterate := func(yield func(T) error) error {
		for t, e := p.Next(); e != io.EOF; t, e = p.Next() {
			if e != nil {
				return e
			}
			e = yield(t)
			if e != nil {
				return e
			}
		}
		return nil
	}
	closef := func() error {
		return CloseAny(p)
	}
	return &Iterator[T]{Iteratef: iterate, Closef: closef}
}

var ErrPull = errors.New("Pull error")

func Pull[T any](it Iter[T], buflen int) *PushPull[T] {
	out := make(chan T, buflen)
	stopc := make(chan struct{})

	go func() {
		it.Iterate(func(t T) error {
			select {
			case out <- t:
				return nil
			case <-stopc:
				return ErrPull
			}
		})
		close(out)
	}()

	yield := func() (T, error) {
		select {
		case t, ok := <-out:
			if ok {
				return t, nil
			}
			return t, ErrPull
		case <-stopc:
			var t T
			return t, ErrPull
		}
	}

	var once sync.Once
	stop := func() error {
		once.Do(func(){ close(stopc) })
		return nil
	}

	return &PushPull[T]{yield, stop}
}

func IntIter(n int) *Iterator[int] {
	return &Iterator[int]{
		Iteratef: func(yield func(int) error) error {
			for i := 0; i < n; i++ {
				err := yield(i)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Closef: func() error { return nil },
	}
}

func SliceIter[T any](s []T) *Iterator[T] {
	return &Iterator[T]{
		Iteratef: func(yield func(T) error) error {
			for _, t := range s {
				err := yield(t)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Closef: func() error { return nil },
	}
}

func SlicePtrIter[T any](s []T) *Iterator[*T] {
	return &Iterator[*T]{
		Iteratef: func(yield func(*T) error) error {
			for i, _ := range s {
				err := yield(&s[i])
				if err != nil {
					return err
				}
			}
			return nil
		},
		Closef: func() error { return nil },
	}
}

type KeyVal[K comparable, V any] struct {
	Key K
	Val V
}

func MapIter[K comparable, V any](m map[K]V) *Iterator[KeyVal[K, V]] {
	return &Iterator[KeyVal[K, V]]{
		Iteratef: func(yield func(KeyVal[K, V]) error) error {
			for k, v := range m {
				err := yield(KeyVal[K, V]{k, v})
				if err != nil {
					return err
				}
			}
			return nil
		},
		Closef: func() error { return nil },
	}
}

func Collect[T any](it Iter[T]) ([]T, error) {
	var out []T
	err := it.Iterate(func(val T) error {
		out = append(out, val)
		return nil
	})
	return out, err
}
