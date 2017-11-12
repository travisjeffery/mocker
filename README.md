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

Example:

Your interface:

``` go
// user_service.go

package user

type UserService interface {
  Get(id string) (*User, error)
}
```

```
$ mocker --out mock/user_service_mock.go --pkg mock . UserService
```

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

## License

MIT

---

- [travisjeffery.com](http://travisjeffery.com)
- GitHub [@travisjeffery](https://github.com/travisjeffery)
- Twitter [@travisjeffery](https://twitter.com/travisjeffery)
- Medium [@travisjeffery](https://medium.com/@travisjeffery)


