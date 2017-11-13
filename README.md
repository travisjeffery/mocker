# mocker

mocker for go.

## Install

``` sh
$ go install github.com/travisjeffery/mocker/cmd/mocker
```

## Usage

``` sh
usage: mocker [<flags>] <src> <ifaces>...

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --out=OUT  File to write mocks to. Stdout by default.
  --pkg=PKG  Name of package for mocks. Inferred by default.

Args:
  <src>     Directory to find interfaces.
  <ifaces>  Interfaces to mock.
```

## Example

Your interface:

``` go
// user_service.go

package user

type UserService interface {
    Get(id string) (*User, error)
}
```

Generate the mock:

```
$ mocker --out mock/user_service_mock.go --pkg mock . UserService
```

Use in your tests:

``` go
// test_endpoint.go

package test_endpoint

func TestUserServiceEndpoint(t *testing.T) {
    us := &mock.UserService{
        GetUserFunc: func(id string) (*user.User, error) {
            return &User{ID: id}, nil
        },
    }
    ep := endpoint.Endpoint{UserService: us}
    resp, err := ep.GetUser(endpoint.GetUserRequest{ID: "travisjeffery"})
    // ...
    if !us.GetUserCalled() {
        t.Error("expected endpoint to call GetUser")
    }
}
```

## API

For each method in the interface, the generated mock struct has methods:

- `__METHOD__(args...) (returns...)`
  Your mocked API which calls the func you instantiated the mock with.

- `__METHOD__Called() bool`
  Returns true if the mocked API was called at least once.

- `__METHOD__Calls() []struct{{args...}}`
  Returns a slice of structs, one struct per call, the struct containing the
  args of the call.

Finally one method to reset all calls on the mock:

- `Reset()`
  Resets the calls made to the mocked APIs.

## License

MIT

---

- [travisjeffery.com](http://travisjeffery.com)
- GitHub [@travisjeffery](https://github.com/travisjeffery)
- Twitter [@travisjeffery](https://twitter.com/travisjeffery)
- Medium [@travisjeffery](https://medium.com/@travisjeffery)


