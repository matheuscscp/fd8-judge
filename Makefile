TESTABLE_PACKAGES := `go list ./... | egrep -v 'mocks|protos|migrations' | grep 'fd8-judge'`
INTERFACES := $(shell find . -name '*interface.go')
MOCKS := $(patsubst %.go,%.go,$(INTERFACES:interface.go=mock.go))

setup:
	@cd ~; go get github.com/golangci/golangci-lint/cmd/golangci-lint
	@cd ~; go get github.com/golang/mock/mockgen

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

test: test-unit test-integration

test-unit:
	@echo "UNIT TESTS"
	@go test ${TESTABLE_PACKAGES} -tags=unit -coverprofile integration.coverprofile
	@echo "\n"

test-integration:
	@echo "INTEGRATION TESTS"
	@go test ${TESTABLE_PACKAGES} -tags=integration -coverprofile integration.coverprofile
	@echo "\n"

mocks: ${MOCKS}

%mock.go: %interface.go
	@mockgen -source $< -package $$(basename $$(dirname "$<")) -destination $@

.PHONY: setup build clean lint check-golangci-lint test test-unit test-integration mocks
