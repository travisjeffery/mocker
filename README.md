# mocker

Mocker for go.

Yeah there's many mockers now but each one I tried failed at least one of:

- Vendored dependencies
- Packages with named imports (e.g. `import mypkg "encoding/json"`)
- Thread safety

This one just works.

## Install

Go:

``` sh
$ go install github.com/travisjeffery/mocker/cmd/mocker
```

Homebrew:

``` sh
$ brew install travisjeffery/homebrew-tap/mocker
```

[Download](https://github.com/travisjeffery/mocker/releases)

## Usage

``` sh
usage: mocker [<flags>] [<src>] [<ifaces>...]

Flags:
  --help             Show context-sensitive help (also try --help-long and
                     --help-man).
  --dst=DST          File write mocks. Leave blank to write to Stdout.
  --pkg=PKG          Name of package for mocks. Inferred by default.
  --prefix="Mock"    Prefix of mock names.
  --suffix=SUFFIX    Suffix of mock names.
  --selfpkg=SELFPKG  The full package import path for the generated code. The
                     purpose of this flag is to prevent import cycles in the
                     generated code by trying to include its own package. This
                     can happen if the mock's package is set to one of its
                     inputs (usually the main one) and the output is stdio so
                     mockgen cannot detect the final output package. Setting
                     this flag will then tell mockgen which import to exclude.

Args:
  [<src>]     File to find interfaces.
  [<ifaces>]  Interfaces to mock.
```

## CLI example

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
$ mocker --dst mock/user_service_mock.go --pkg mock user_service.go UserService
```

Use in your tests:

``` go
// test_endpoint.go

package test_endpoint

func TestUserServiceEndpoint(t *testing.T) {
    us := &mock.UserService{
        GetFunc: func(id string) (*user.User, error) {
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

## Go generate example

``` go
// user_service.go

package user

//go:generate mocker --dst mock/$GOFILE --pkg mock $GOFILE UserService
type UserService interface {
	Get(id string) (*User, error)
}
```

Generate the mock:

```
$ go generate
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
