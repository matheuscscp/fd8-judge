
# prepend GOFLAGS env var with '-mod=vendor' so all go commands use vendor folder
GOFLAGS := -mod=vendor $(GOFLAGS)
export GOFLAGS

# ==================================================================================================
# long shell commands
# ==================================================================================================

MOCKGEN := go run github.com/golang/mock/mockgen
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/cmd/golangci-lint
GOCOVMERGE := go run github.com/wadey/gocovmerge

# ==================================================================================================
# build and clean
# ==================================================================================================

BUILD_TARGETS := bin/fd8-judge

.PHONY: build
build: $(BUILD_TARGETS)

bin/fd8-judge:
	go build -o $@

.PHONY: clean
clean:
	rm -rf $(BUILD_TARGETS) bin/ cov/

# ==================================================================================================
# gen and clean-gen (artifacts generation: mocks, protos...)
# ==================================================================================================

INTERFACES := `grep -rls ./pkg ./judge -e 'interface {$$'`
MOCKS := $(shell echo $(INTERFACES) | sed 's/\.\//\.\/test\/mocks\//g')

.PHONY: gen
gen: $(MOCKS)

./test/mocks/%: ./%
	$(MOCKGEN) -source $< -package $$(basename $$(dirname "$<")) -destination $@

.PHONY: clean-gen
clean-gen:
	rm -rf $(MOCKS)

# ==================================================================================================
# fix, lint, test and cover
# ==================================================================================================

TESTABLE_PACKAGES := `go list ./... | egrep -v 'protos|migrations|test|cmd|tools' | grep 'fd8-judge/'`
COVERAGE_FILES := cov/unit.out cov/integration.out

.PHONY: fix
fix:
	go mod tidy
	go mod vendor

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit: cov
	go test $(TESTABLE_PACKAGES) -tags=unit -coverprofile cov/unit.out

.PHONY: test-integration
test-integration: cov
	go test $(TESTABLE_PACKAGES) -tags=integration -coverprofile cov/integration.out

cov:
	mkdir -p ./cov

.PHONY: cover
cover: cov/full.out
	go tool cover -func=cov/full.out | grep total | awk '{print $$3}'

cov/full.out: cov $(COVERAGE_FILES)
	$(GOCOVMERGE) $(COVERAGE_FILES) > $@
