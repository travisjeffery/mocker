generate:
	go run cmd/mocker/main.go --out test/out.go test Iface

test:
	go test -v ./test

.PHONY: test generate
