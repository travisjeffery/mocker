.PHONY: clean
clean:
	rm -f test/out.go

.PHONY: generate
generate: clean
	go run cmd/mocker/main.go --dst test/out.go test/in.go Iface

.PHONY: test
test:
	go test -v ./test

.PHONY: install
install:
	go install ./cmd/mocker
