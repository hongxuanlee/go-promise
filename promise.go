package promise

import (
	"reflect"
)

type promiseCallback func(interface{}) (interface{}, error)

type errorCallback func(error) (interface{}, error)

// ResolveHandler excute when promise object success
type ResolveHandler func(interface{})

// RejectHandler excute when promise object failed
type RejectHandler func(error)

//Interface represents a Promise Object.
type Interface interface {

	//Then represents a returns a Promise. It takes up to two arguments: callback functions for the success and failure cases of the Promise
	Then(promiseCallback) Interface

	//Catch returns a Promise and deals with rejected cases only
	Catch(errorCallback) Interface

	//Finally returns a Promise. When the promise is settled, whether fulfilled or rejected, the specified callback function is executed. This provides a way for code that must be executed once the Promise has been dealt with to be run whether the promise was fulfilled successfully or rejected
	Finally(func() error) Interface
}

type innerPromise struct {
	done   chan struct{}
	cancel chan struct{}
	val    interface{}
	err    error
}

func transFunc(fn func(...interface{}) (interface{}, error)) func(interface{}) (interface{}, error) {
	return func(v interface{}) (interface{}, error) {
		rt := reflect.TypeOf(v)
		switch rt.Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(v)
			args := make([]interface{}, s.Len())
			for i := 0; i < s.Len(); i++ {
				args[i] = s.Index(i).Interface()
			}
			return fn(args...)
		default:
			return fn(v)
		}
	}
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
func Promisify(fn func(interface{}) (interface{}, error)) func(interface{}) Interface {
	doneChan := make(chan struct{})
	cancelChan := make(chan struct{})
	promise := innerPromise{doneChan, cancelChan, nil, nil}
	return func(v interface{}) Interface {
		go func() {
			res, err := fn(v)
			if err != nil {
				promise.reject(err)
			} else {
				promise.resolve(res)
			}
		}()
		return &promise
	}
}

func (p *innerPromise) resolve(res interface{}) {
	p.val = res
	p.done <- struct{}{}
}

func (p *innerPromise) reject(err error) {
	p.err = err
	p.cancel <- struct{}{}
}

// Catch when error occurring
func (p *innerPromise) Catch(errCb errorCallback) Interface {
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
func (p *innerPromise) Then(next promiseCallback) Interface {
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

// All returns a single Promise that resolves when all of the promises in the iterable argument have resolved or when the iterable argument contains no promises. It rejects with the reason of the first promise that rejects.
//func All() Interface {
//
//}

// Race returns a promise that resolves or rejects as soon as one of the promises in the iterable resolves or rejects, with the value or reason from that promise.
//func Race() Interface {
//
//}

// Resolve returns a Promise object that is resolved with the given value. If the value is a promise, that promise is returned; if the value is a thenable (i.e. has a "then" method), the returned promise will "follow" that thenable, adopting its eventual state; otherwise the returned promise will be fulfilled with the value.
//func Resolve() Interface {
//
//}

// Reject returns a Promise object that is rejected with the given reason.
//func Reject() Interface {
//
//}
