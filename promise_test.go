package promise

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func addone(num interface{}) (interface{}, error) {
	time.Sleep(1 * time.Second)
	return num.(int) + 1, nil
}

func Test_promise(t *testing.T) {
	p := Promisify(addone)
	var wg sync.WaitGroup
	wg.Add(1)
	p(1).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 2, data, "")
		res = data
		return p(data), nil
	}).Then(func(data interface{}) (res interface{}, err error) {
		assertEqual(t, 3, data, "")
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
