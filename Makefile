APP?=homer
GOOS?=linux

COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
RELEASE?=$(shell cat VERSION)

.PHONY: check
check: prepare_metalinter
	gometalinter --skip vendor --config=./.gometalinter.json ./...

.PHONY: build
build: clean
	CGO_ENABLED=0 GOOS=${GOOS} go build \
		-ldflags "-X main.versionRELEASE=${RELEASE} -X main.versionCOMMIT=${COMMIT} -X main.versionBUILD=${BUILD_TIME}" \
		-o bin/${GOOS}/${APP} 

.PHONY: clean
clean:
	@rm -f bin/${GOOS}/${APP}

.PHONY: vendor
vendor: prepare_glide
	glide update

.PHONY: fmt
fmt:
	go fmt *.go
	go fmt shutdown/*.go
	go fmt stat/*.go
	go fmt database/*.go
	go fmt helper/*.go

HAS_DEP := $(shell command -v dep;)
HAS_GLIDE := $(shell command -v dep;)
HAS_METALINTER := $(shell command -v gometalinter;)

.PHONY: prepare_dep
prepare_dep:
ifndef HAS_DEP
	go get -u -v -d github.com/golang/dep/cmd/dep && \
	go install -v github.com/golang/dep/cmd/dep
endif

.PHONY: prepare_metalinter
prepare_metalinter:
ifndef HAS_METALINTER
	go get -u -v -d github.com/alecthomas/gometalinter && \
	go install -v github.com/alecthomas/gometalinter && \
	gometalinter --install --update
endif
