## Go Promise Libaray

Inspired by JS Standard built-in objects - Promise, This is a Promise Libaray for Golang.

## Promise

A Promise is a proxy for a value not necessarily known when the promise is created. It allows you to associate handlers with an asynchronous action's eventual success value or failure reason. This lets asynchronous methods return values like synchronous methods: instead of immediately returning the final value, the asynchronous method returns a promise to supply the value at some point in the future.

A Promise is in one of these states:

- pending: initial state, neither fulfilled nor rejected.

- fulfilled: meaning that the operation completed successfully.

- rejected: meaning that the operation failed.
A pending promise can either be fulfilled with a value, or rejected with a reason (error). When either of these options happens, the associated handlers queued up by a promise's then method are called. 

### Usage

- Create Promise Function

```go
  p := NewPromise(func(resolve ResolveHandler, reject RejectHandler) {
      time.Sleep(200 * time.Millisecond)
      resolve(100)
  })
  fn.Then(func(data interface{}) (res interface{}, err error) {
      fmt.Println(res)
  })
```

- Promisefy normal function

```go
  import "github.com/hongxuanlee/go-promise"
  
  func addone(num interface{}) (interface{}, error) {
      time.Sleep(200 * time.Millisecond)
      return num.(int) + 1, nil
  }
  p := promise.Promisify(addone)
  var wg sync.WaitGroup
  wg.Add(1)
  p(1).Then(func(data interface{}) (res interface{}, err error) {
       return p(data), nil
  }).Then(func(data interface{}) (res interface{}, err error) {
      fmt.Println(data) // should be 3
      wg.Done()
      return
  })
  
```

- Error catch 

```go
  import "github.com/hongxuanlee/go-promise"
  
  fn := promise.NewPromise(func(resolve ResolveHandler, reject RejectHandler) {
  	time.Sleep(200 * time.Millisecond)
  	reject(errors.New("something wrong"))
  })
  var wg sync.WaitGroup
  wg.Add(1)
  fn.Then(func(data interface{}) (res interface{}, err error) {
    // not excute this block
  	return
  }).Catch(func(e error) (res interface{}, err error) {
    fmt.Println(e) // error catch here 
    // handleError(e)
    wg.Done()
    return
  })
  wg.Wait()
```

- Promise All

```go
  import "github.com/hongxuanlee/go-promise"
  
  ps := []Interface{p1, p2, p3}
  var wg sync.WaitGroup
  wg.Add(1)
  promise.All(ps).Then(func(data interface{}) (res interface{}, err error) {
      fmt.Println(data) 
      wg.Done()
      return
  })
  wg.Wait()

```

- Promise Race

```go
  import "github.com/hongxuanlee/go-promise"
  
  ps := []Interface{p1, p2, p3}
  var wg sync.WaitGroup
  wg.Add(1)
  promise.Race(ps).Then(func(data interface{}) (res interface{}, err error) {
      fmt.Println(data) 
      wg.Done()
      return
  })
  wg.Wait()

```








