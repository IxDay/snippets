GOPATH = $(abspath $(CURDIR)/../..)

.PHONY: run
run:
	@GOPATH=$(GOPATH) go run cmd/*.go

.PHONY: test
test:
	GOPATH=$(GOPATH) go test -v -cover -coverprofile coverage.txt
	GOPATH=$(GOPATH) go tool cover -func coverage.txt
	@rm coverage.txt
	@go vet *.go
	@bash -c "diff -u <(echo -n) <(gofmt -d ./)"
