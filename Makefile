default: ci

ci: lint test fmt-check imports-check integration

GOLANGCILINTVERSION?=1.49.0
GOIMPORTSVERSION?=v0.1.12
GOXVERSION?=v1.0.1
GOTESTSUMVERSION?=v1.8.2
GOREVIVEVERSION?=v1.2.3
GOLANGCILINTLSVERSION?=v0.0.7

CIARTIFACTS?=ci-artifacts
COVERAGEOUT?=coverage.out
COVERAGEHTML?=coverage.html
PACKAGENAME?=ecomm-reporter
CLINAME?=ecomm-rpt
GOFLAGS=-mod=vendor
CGO_ENABLED?=1
GO_LDFLAGS="-X github.com/ipcrm/sko-hol-ssrf/cli/cmd.Version=$(shell cat VERSION) \
            -X github.com/ipcrm/sko-hol-ssrf/cli/cmd.GitSHA=$(shell git rev-parse HEAD) \
            -X github.com/ipcrm/sko-hol-ssrf/cli/cmd.BuildTime=$(shell date +%Y%m%d%H%M%S)"

export GOFLAGS GO_LDFLAGS CGO_ENABLED GOX_LINUX_AMD64_LDFLAGS

.PHONY: help
help:
	@echo "-------------------------------------------------------------------"
	@echo "Makefile helper:"
	@echo ""
	@grep -Fh "##" $(MAKEFILE_LIST) | grep -v grep | sed -e 's/\\$$//' | sed -E 's/^([^:]*):.*##(.*)/ \1 -\2/'
	@echo "-------------------------------------------------------------------"

.PHONY: prepare
prepare: install-tools go-vendor ## Initialize the go environment

.PHONY: test
test: prepare ## Run all tests
	CI=true gotestsum -f testname -- -v -cover -coverprofile=$(COVERAGEOUT) $(shell go list ./... | grep -v integration)

.PHONY: coverage
coverage: test ## Output coverage profile information for each function
	go tool cover -func=$(COVERAGEOUT)

.PHONY: coverage-html
coverage-html: test ## Generate HTML representation of coverage profile
	go tool cover -html=$(COVERAGEOUT)

.PHONY: go-vendor
go-vendor: ## Runs go mod tidy, vendor and verify to cleanup, copy and verify dependencies
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: lint
lint: ## Runs go linter
	golangci-lint run

.PHONY: fmt
fmt: ## Runs and applies go formatting changes
	@gofmt -w -l $(shell go list -f {{.Dir}} ./...)
	@goimports -w -l $(shell go list -f {{.Dir}} ./...)

.PHONY: fmt-check
fmt-check: ## Lists formatting issues
	@test -z $(shell gofmt -l $(shell go list -f {{.Dir}} ./...))

.PHONY: imports-check
imports-check: ## Lists imports issues
	@test -z $(shell goimports -l $(shell go list -f {{.Dir}} ./...))

.PHONY: build-cli-cross-platform
build-cli-cross-platform:
	gox -output="bin/$(PACKAGENAME)-backend-{{.OS}}-{{.Arch}}" \
            -os="linux" \
            -arch="amd64 386" \
            -osarch="darwin/amd64 darwin/arm64 linux/arm linux/arm64" \
            -ldflags=$(GO_LDFLAGS) \
            github.com/ipcrm/sko-hol-ssrf/cli/backend
	gox -output="bin/$(PACKAGENAME)-frontend-{{.OS}}-{{.Arch}}" \
            -os="linux" \
            -arch="amd64 386" \
            -osarch="darwin/amd64 darwin/arm64 linux/arm linux/arm64" \
            -ldflags=$(GO_LDFLAGS) \
            github.com/ipcrm/sko-hol-ssrf/cli/frontend

.PHONY: build-cli-dev
build-cli-dev:
ifeq (x86_64, $(shell uname -m))
	gox -output="bin/$(PACKAGENAME)-backend-{{.OS}}-{{.Arch}}" \
						-os=$(shell uname -s | tr '[:upper:]' '[:lower:]') \
						-arch="amd64" \
						-gcflags="all=-N -l" \
						-ldflags=$(GO_LDFLAGS) \
						github.com/ipcrm/sko-hol-ssrf/cli/backend
else
	gox -output="bin/$(PACKAGENAME)-backend-{{.OS}}-{{.Arch}}" \
						-os=$(shell uname -s | tr '[:upper:]' '[:lower:]') \
						-arch="386" \
						-gcflags="all=-N -l" \
						-osarch="$(shell uname -s | tr '[:upper:]' '[:lower:]')/amd $(shell uname -s | tr '[:upper:]' '[:lower:]')/arm" \
						-ldflags=$(GO_LDFLAGS) \
						github.com/ipcrm/sko-hol-ssrf/cli/backend
endif
ifeq (x86_64, $(shell uname -m))
	gox -output="bin/$(PACKAGENAME)-frontend-{{.OS}}-{{.Arch}}" \
						-os=$(shell uname -s | tr '[:upper:]' '[:lower:]') \
						-arch="amd64" \
						-gcflags="all=-N -l" \
						-ldflags=$(GO_LDFLAGS) \
						github.com/ipcrm/sko-hol-ssrf/cli/frontend
else
	gox -output="bin/$(PACKAGENAME)-backend-{{.OS}}-{{.Arch}}" \
						-os=$(shell uname -s | tr '[:upper:]' '[:lower:]') \
						-arch="386" \
						-gcflags="all=-N -l" \
						-osarch="$(shell uname -s | tr '[:upper:]' '[:lower:]')/amd $(shell uname -s | tr '[:upper:]' '[:lower:]')/arm" \
						-ldflags=$(GO_LDFLAGS) \
						github.com/ipcrm/sko-hol-ssrf/cli/frontend
endif

.PHONY: copy-bins
copy-bins:
ifeq (x86_64, $(shell uname -m))
	cp bin/$(PACKAGENAME)-backend-$(shell uname -s | tr '[:upper:]' '[:lower:]')-amd64 /usr/local/bin/$(CLINAME)-be
else
	cp bin/$(PACKAGENAME)-backend-$(shell uname -s | tr '[:upper:]' '[:lower:]')-386 /usr/local/bin/$(CLINAME)-be
endif
	@echo "\nThe backend cli has been installed at /usr/local/bin"

ifeq (x86_64, $(shell uname -m))
	cp bin/$(PACKAGENAME)-frontend-$(shell uname -s | tr '[:upper:]' '[:lower:]')-amd64 /usr/local/bin/$(CLINAME)-fe
else
	cp bin/$(PACKAGENAME)-frontend-$(shell uname -s | tr '[:upper:]' '[:lower:]')-386 /usr/local/bin/$(CLINAME)-fe
endif
	@echo "\nThe frontend cli has been installed at /usr/local/bin"

.PHONY: install-cli-dev
install-cli-dev: build-cli-dev copy-bins

.PHONY: install-cli
install-cli: build-cli-cross-platform copy-bins

.PHONY: build-all-dev
build-all-dev: install-cli-dev

.PHONY: integration-test
integration-test: install-tools ## Run integration tests
	PATH=$(PWD)/bin:${PATH} gotestsum -f testname -- -v github.com/ipcrm/sko-hol-ssrf/test/integration

.PHONY: dev-docs
dev-docs:
	cd docs && yarn && yarn start

.PHONY: install-tools
install-tools: ## Install go indirect dependencies
ifeq (, $(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v$(GOLANGCILINTVERSION)
endif
ifeq (, $(shell which goimports))
	GOFLAGS=-mod=readonly go install golang.org/x/tools/cmd/goimports@$(GOIMPORTSVERSION)
endif
ifeq (, $(shell which gox))
	GOFLAGS=-mod=readonly go install github.com/mitchellh/gox@$(GOXVERSION)
endif
ifeq (, $(shell which gotestsum))
	GOFLAGS=-mod=readonly go install gotest.tools/gotestsum@$(GOTESTSUMVERSION)
endif
ifeq (, $(shell which revive))
	GOFLAGS=-mod=readonly go install github.com/mgechev/revive@$(GOREVIVEVERSION)
endif
ifeq (, $(shell which golangci-lint-langserver))
	GOFLAGS=-mod=readonly go install github.com/nametake/golangci-lint-langserver@$(GOLANGCILINTLSVERSION)
endif

.PHONY: release
release: prepare
	scripts/release.sh prepare
