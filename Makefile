.PHONY: clean
clean:
	rm -f test/out.go

.PHONY: generate
generate: clean
	go run cmd/mocker/main.go --out test/out.go test Iface

.PHONY: test
test:
	go test -v ./test

