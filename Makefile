TESTABLE_PACKAGES = `go list ./... | egrep -v 'mocks|protos|migrations' | grep 'fd8-judge/'`

build:
	@go build -o fd8-judge

clean:
	@rm -rf fd8-judge *.coverprofile

test-integration:
	@go test ${TESTABLE_PACKAGES} -tags=integration -coverprofile integration.coverprofile

test: test-integration

.PHONY: build clean test test-integration
