TESTABLE_PACKAGES := `go list ./... | egrep -v 'protos|migrations|test' | grep 'fd8-judge/'`
INTERFACES := `grep -rls ./pkg/bll ./pkg/dal ./pkg/services -e 'interface {$$'`
MOCKS := $(shell echo ${INTERFACES} | sed 's/pkg/test\/mocks\/pkg/g')

setup:
	@if ! which golangci-lint > /dev/null; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $$(go env GOPATH)/bin v1.21.0 \
		exit 1; \
	fi

build:
	@go build -o fd8-judge

clean:
	@rm -rf fd8-judge *coverage.out

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
	@go test ${TESTABLE_PACKAGES} -tags=unit -coverprofile unit.coverage.out
	@echo "\n"

test-integration:
	@echo "INTEGRATION TESTS"
	@go test ${TESTABLE_PACKAGES} -tags=integration -coverprofile integration.coverage.out
	@echo "\n"

mocks: ${MOCKS}

./test/mocks/%: ./%
	@go run github.com/golang/mock/mockgen -source $< -package $$(basename $$(dirname "$<")) -destination $@

cover:
	@go run github.com/wadey/gocovmerge unit.coverage.out integration.coverage.out > full.coverage.out
	@go tool cover -func=full.coverage.out | grep total | awk '{print $$3}'

.PHONY: setup build clean lint check-golangci-lint test test-unit test-integration mocks cover
