TESTABLE_PACKAGES = `go list ./... | egrep -v 'mocks|protos|migrations' | grep 'fd8-judge/'`

setup:
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint

build:
	@go build -o fd8-judge

clean:
	@rm -rf fd8-judge *.coverprofile

lint: check-golangci-lint
	@golangci-lint run

check-golangci-lint:
	@if ! which golangci-lint > /dev/null; then \
		echo -e 'Please install golangci-lint running make setup. See https://github.com/golangci/golangci-lint#local-installation'; \
		exit 1; \
	fi

test-integration:
	@go test ${TESTABLE_PACKAGES} -tags=integration -coverprofile integration.coverprofile

test: test-integration

.PHONY: setup build clean lint check-golangci-lint test test-integration
