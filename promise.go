package promise

import (
	"reflect"
)

type promiseCallback func(interface{}) (interface{}, error)

//Interface represents a Promise Object.
type Interface interface {

	//Then represents a returns a Promise. It takes up to two arguments: callback functions for the success and failure cases of the Promise
	Then(promiseCallback) Interface

	//Catch returns a Promise and deals with rejected cases only
	//Catch(func(error) (Interface, error))

	//Finally returns a Promise. When the promise is settled, whether fulfilled or rejected, the specified callback function is executed. This provides a way for code that must be executed once the Promise has been dealt with to be run whether the promise was fulfilled successfully or rejected
	//	Finally(func() error)
}

type innerPromise struct {
	done   chan struct{}
	cancel chan struct{}
	val    interface{}
	err    error
}

// NewPromise represent initial promise object
func NewPromise(fn func(func(interface{}), func(error))) Interface {
	p := innerPromise{make(chan struct{}), make(chan struct{}), nil, nil}
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
			promise.done <- struct{}{}
			promise.val = res
			promise.err = err
		}()
		return &promise
	}
}

//func (p *Promise) Catch(handle promiseCallback) Interface {
//}

// Then returns a Promise. callback functions.
func (p *innerPromise) Then(next promiseCallback) Interface {
	var result interface{}
	nextDone := make(chan struct{})
	nextCancel := make(chan struct{})
	nextPromise := innerPromise{nextDone, nextCancel, nil, nil}
	select {
	case <-p.done:
		result = p.val
	case <-p.cancel:
		return nil
	}
	if p.err == nil {
		nextRes, err := next(result)
		if err == nil && nextRes != nil {
			if reflect.TypeOf(nextRes).String() == "*promise.innerPromise" {
				return nextRes.(*innerPromise)
			}
			nextPromise.val = nextRes
		}
	}
	return &nextPromise
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
