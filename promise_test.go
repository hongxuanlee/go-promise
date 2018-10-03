package promise

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func addone(num interface{}) (interface{}, error) {
	time.Sleep(200 * time.Millisecond)
	return num.(int) + 1, nil
}

func Test_promise(t *testing.T) {
	p := Promisify(addone)
	var wg sync.WaitGroup
	wg.Add(1)
	p(1).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 2, data, "")
		return p(data), nil
	}).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 3, data, "")
		wg.Done()
		return
	})
	wg.Wait()
}

func Test_newPromise(t *testing.T) {
	fn := NewPromise(func(resolve ResolveHandler, reject RejectHandler) {
		time.Sleep(200 * time.Millisecond)
		resolve(100)
	})
	var wg sync.WaitGroup
	wg.Add(1)
	fn.Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 100, data, "")
		wg.Done()
		return
	})
	wg.Wait()
}

func Test_reject(t *testing.T) {
	fn := NewPromise(func(resolve ResolveHandler, reject RejectHandler) {
		time.Sleep(200 * time.Millisecond)
		reject(errors.New("something wrong"))
	})
	var wg sync.WaitGroup
	wg.Add(1)
	fn.Then(func(data interface{}) (res interface{}, err error) {
		t.Fatal("should catch error")
		wg.Done()
		return
	}).Catch(func(e error) (res interface{}, err error) {
		wg.Done()
		return
	})
	wg.Wait()
}

func Test_errorReject(t *testing.T) {
	p := Promisify(addone)
	var wg sync.WaitGroup
	wg.Add(1)
	p(1).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 2, data, "")
		err = errors.New("something wrong")
		return
	}).Then(func(data interface{}) (res interface{}, err error) {
		t.Fatal("should catch error")
		wg.Done()
		return
	}).Catch(func(e error) (res interface{}, err error) {
		wg.Done()
		return
	})
	wg.Wait()
}

func Test_Finally(t *testing.T) {
	p := Promisify(addone)
	var wg sync.WaitGroup
	wg.Add(1)
	p(1).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 2, data, "")
		return p(data), nil
	}).Finally(func() (err error) {
		err = errors.New("something wrong")
		return
	}).Catch(func(e error) (interface{}, error) {
		wg.Done()
		return nil, nil
	})
	wg.Wait()
}

func Test_promise_lazy_excute(t *testing.T) {
	p := Promisify(addone)
	var wg sync.WaitGroup
	wg.Add(1)
	fn := p(1)
	time.Sleep(1 * time.Second)
	fn.Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 2, data, "")
		return p(data), nil
	}).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 3, data, "")
		wg.Done()
		return
	})
	wg.Wait()
}

func Test_PromiseAll(t *testing.T) {
	p1 := Resolve(1)
	p2 := Resolve(2)
	p3 := Resolve(3)

	ps := []Interface{p1, p2, p3}
	var wg sync.WaitGroup
	wg.Add(1)
	All(ps).Then(func(data interface{}) (res interface{}, err error) {
		wg.Done()
		return
	})
	wg.Wait()
}

func Test_PromiseRace(t *testing.T) {
	p1 := Resolve(1)
	p := Promisify(addone)
	p2 := p(2)
	p3 := p(3)

	ps := []Interface{p1, p2, p3}
	var wg sync.WaitGroup
	wg.Add(1)
	Race(ps).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 1, data, "")
		wg.Done()
		return
	})
	wg.Wait()
}

func assertEqual(t *testing.T, expect interface{}, actual interface{}, message string) {
	if expect == actual {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("expect %v !=  actual %v", expect, actual)
	}
	t.Fatal(message)
}
