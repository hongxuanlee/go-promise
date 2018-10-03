package promise

import (
	"reflect"
)

// ResolveHandler excute when promise object success
type ResolveHandler func(interface{})

// RejectHandler excute when promise object failed
type RejectHandler func(error)

// PromiseCallback excute when promise fulfill success
type PromiseCallback func(interface{}) (interface{}, error)

// ErrorCallback excute when promise fullfill fail
type ErrorCallback func(error) (interface{}, error)

//Interface represents a Promise Object.
type Interface interface {

	//Then represents a returns a Promise. It takes up to two arguments: callback functions for the success and failure cases of the Promise
	Then(PromiseCallback) Interface

	//Catch returns a Promise and deals with rejected cases only
	Catch(ErrorCallback) Interface

	//Finally returns a Promise. When the promise is settled, whether fulfilled or rejected, the specified callback function is executed. This provides a way for code that must be executed once the Promise has been dealt with to be run whether the promise was fulfilled successfully or rejected
	Finally(func() error) Interface
}

type innerPromise struct {
	done   chan struct{}
	cancel chan struct{}
	val    interface{}
	err    error
}

// NewPromise represent initial promise object
func NewPromise(fn func(ResolveHandler, RejectHandler)) Interface {
	p := innerPromise{make(chan struct{}), make(chan struct{}), nil, nil}
	go func() {
		fn(func(res interface{}) {
			p.resolve(res)
		}, func(err error) {
			p.reject(err)
		})
	}()
	return &p
}

// Promisify represents to promisefy sync function to async
func Promisify(fn func(v ...interface{}) (interface{}, error)) func(...interface{}) Interface {
	doneChan := make(chan struct{})
	cancelChan := make(chan struct{})
	promise := innerPromise{doneChan, cancelChan, nil, nil}
	return func(v ...interface{}) Interface {
		go func() {
			res, err := fn(v...)
			if err != nil {
				promise.reject(err)
			} else {
				promise.resolve(res)
			}
		}()
		return &promise
	}
}

// Catch when error occurring
func (p *innerPromise) Catch(errCb ErrorCallback) Interface {
	if p.err == nil {
		return p
	}
	nextp := innerPromise{make(chan struct{}), make(chan struct{}), nil, nil}
	nextRes, err := errCb(p.err)
	if err == nil && nextRes != nil {
		if reflect.TypeOf(nextRes).String() == "*promise.innerPromise" {
			return nextRes.(*innerPromise)
		}
		nextp.val = nextRes
	}
	return &nextp
}

// Then returns a Promise. callback functions.
func (p *innerPromise) Then(next PromiseCallback) Interface {
	if p.err != nil {
		return p
	}
	var result interface{}
	nextDone := make(chan struct{})
	nextCancel := make(chan struct{})
	nextPromise := innerPromise{nextDone, nextCancel, nil, nil}
	select {
	case <-p.done:
		result = p.val
	case <-p.cancel:
		nextPromise.err = p.err
		return &nextPromise
	}
	nextRes, err := next(result)
	if err == nil {
		if nextRes != nil && reflect.TypeOf(nextRes).String() == "*promise.innerPromise" {
			return nextRes.(*innerPromise)
		}
		nextPromise.val = nextRes
	} else {
		nextPromise.err = err
	}

	return &nextPromise
}

func (p *innerPromise) Finally(fn func() error) Interface {
	err := fn()
	if err != nil {
		p.err = err
	}
	return p
}

func (p *innerPromise) resolve(res interface{}) {
	p.val = res
	p.done <- struct{}{}
}

func (p *innerPromise) reject(err error) {
	p.err = err
	p.cancel <- struct{}{}
}

// All returns a single Promise that resolves when all of the promises in the iterable argument have resolved or when the iterable argument contains no promises. It rejects with the reason of the first promise that rejects.
func All(promises []Interface) Interface {
	nextp := innerPromise{make(chan struct{}), make(chan struct{}), nil, nil}
	result := make([]interface{}, len(promises))
	count := 0
	for i, p := range promises {
		go func(i int, p Interface) {
			p.Then(func(data interface{}) (res interface{}, err error) {
				count++
				result[i] = data
				if count == len(promises) {
					nextp.resolve(result)
					return
				}
				return
			}).Catch(func(e error) (res interface{}, err error) {
				if e != nil {
					nextp.reject(e)
				}
				return
			})
		}(i, p)
	}

	return &nextp
}

// Race returns a promise that resolves or rejects as soon as one of the promises in the iterable resolves or rejects, with the value or reason from that promise.
func Race(promises []Interface) Interface {
	nextp := innerPromise{make(chan struct{}), make(chan struct{}), nil, nil}
	for i, p := range promises {
		go func(i int, p Interface) {
			p.Then(func(data interface{}) (res interface{}, err error) {
				nextp.resolve(data)
				return
			}).Catch(func(e error) (res interface{}, err error) {
				if e != nil {
					nextp.reject(e)
				}
				return
			})
		}(i, p)
	}
	return &nextp
}

// Resolve returns a Promise object that is resolved with the given value. If the value is a promise, that promise is returned; if the value is a thenable (i.e. has a "then" method), the returned promise will "follow" that thenable, adopting its eventual state; otherwise the returned promise will be fulfilled with the value.
func Resolve(res interface{}) Interface {
	p := innerPromise{make(chan struct{}), make(chan struct{}), nil, nil}
	go func() {
		p.resolve(res)
	}()
	return &p
}

// Reject returns a Promise object that is rejected with the given reason.
func Reject(err error) Interface {
	p := innerPromise{make(chan struct{}), make(chan struct{}), nil, nil}
	go func() {
		p.reject(err)
	}()
	return &p
}
