
# prepend GOFLAGS env var with '-mod=vendor' so all go commands use vendor folder
ifeq (,$(findstring -mod=vendor,$(GOFLAGS)))
	GOFLAGS := -mod=vendor $(GOFLAGS)
endif
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

SOURCE_FILES := $(shell find . -name '*.go' | egrep -v 'test|vendor|tools')
BUILD_TARGETS := bin/fd8-judge

.PHONY: build
build: $(BUILD_TARGETS)

bin/fd8-judge: $(SOURCE_FILES)
	go build -race -o $@

.PHONY: clean
clean:
	rm -rf bin/ cov/ $(BUILD_TARGETS)

# ==================================================================================================
# gen and clean-gen (artifacts generation: mocks, protos...)
# ==================================================================================================

INTERFACES := `grep -rls ./pkg ./judge -e 'interface {$$'`
MOCKS := $(shell echo $(INTERFACES) | sed 's/\.\//\.\/test\/mocks\/gen\//g')

.PHONY: gen
gen: $(MOCKS)

./test/mocks/gen/%: ./%
	$(MOCKGEN) -source $< -package $$(basename $$(dirname "$<")) -destination $@

.PHONY: clean-gen
clean-gen:
	rm -rf $(MOCKS)

# ==================================================================================================
# fix, lint, test and coverage
# ==================================================================================================

FILTER_TESTABLE_PACKAGES := egrep -v 'proto|migrations|test|cmd|tools'
TESTABLE_PACKAGES := $(shell go list ./... | $(FILTER_TESTABLE_PACKAGES) | grep 'fd8-judge/')

UNIT_COVERAGE_FILE := cov/unit.out
INTEGRATION_COVERAGE_FILE := cov/integration.out
COVERAGE_FILES := $(UNIT_COVERAGE_FILE) $(INTEGRATION_COVERAGE_FILE)

TEST_GARBAGE_FILES := judge/serverFiles/ judge/bundle/ judge/interactor* judge/solution* judge/outputs*

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
test-unit: clean-test cov
	go test $(TESTABLE_PACKAGES) -race -coverpkg=./... -tags=unit -coverprofile=$(UNIT_COVERAGE_FILE)

.PHONY: test-integration
test-integration: clean-test cov bin/fd8-judge
	go test $(TESTABLE_PACKAGES) -race -coverpkg=./... -p 1 -tags=integration -coverprofile=$(INTEGRATION_COVERAGE_FILE)

.PHONY: clean-test
clean-test:
	rm -rf $(TEST_GARBAGE_FILES)

cov:
	mkdir -p ./cov

cov/coverage.txt: cov $(COVERAGE_FILES)
	$(GOCOVMERGE) $(COVERAGE_FILES) | $(FILTER_TESTABLE_PACKAGES) > $@
	@scripts/cover-percentage

.PHONY: cover
cover: cov/coverage.txt
	@cat cov/percentage.out

# ==================================================================================================
# all
# ==================================================================================================

.PHONY: all
all: gen fix lint test cov/coverage.txt
