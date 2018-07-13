GOPATH = $(abspath $(CURDIR)/../..)

.PHONY: run
run:
	@GOPATH=$(GOPATH) go run cmd/*.go

.PHONY: fmt
fmt:
	@bash -c "diff -u <(echo -n) <(gofmt -d ./)"

.PHONY: vet
vet:
	@go vet *.go

.PHONY: cover
cover:
	GOPATH=$(GOPATH) go test -v -cover -coverprofile coverage.txt
	GOPATH=$(GOPATH) go tool cover -func coverage.txt
	@rm coverage.txt

.PHONY: test
test: cover fmt vet
