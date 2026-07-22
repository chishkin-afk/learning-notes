package app

import (
	"context"
	"errors"
	"slices"
	"sync"
)

type closeFn func(ctx context.Context) error

type closer struct {
	mu    sync.Mutex
	funcs []closeFn
}

func (c *closer) all(ctx context.Context) error {
	c.mu.Lock()
	funcs := slices.Clone(c.funcs)
	c.funcs = nil
	c.mu.Unlock()

	errs := make([]error, 0)
	for _, fn := range slices.Backward(funcs) {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (c *closer) add(fn closeFn) {
	c.mu.Lock()
	c.funcs = append(c.funcs, fn)
	c.mu.Unlock()
}

var globalCloser closer

// All closes all the deps
//
// It iterates sequentially through all dependencies,
// closing them.
func All(ctx context.Context) error {
	return globalCloser.all(ctx)
}

// Add adds new close function
//
// Adds a new dependency closure function to the list
// can be called asynchronously.
func Add(fn closeFn) {
	globalCloser.add(fn)
}
