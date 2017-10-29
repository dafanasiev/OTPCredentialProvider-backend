LDFLAGS += -X "main.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
#LDFLAGS += -X "main.BuildGitHash=$(shell git rev-parse HEAD)"
NOW = $(shell date -u '+%Y%m%d%I%M%S')

TAGS = ""
BUILD_FLAGS = "-v"

HAS_DEP := $(shell command -v dep;)

all: build

vendoring: prepare_dep
	dep ensure

build: vendoring
	go build -o bin/checkserver $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' github.com/dafanasiev/OTPCredentialProvider-backend/app/checkserver
	go build -o bin/showqr $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' github.com/dafanasiev/OTPCredentialProvider-backend/app/showqr

clean:
	rm -rf bin/
	rm -rf vendor/
	rm -rf vendor.orig/

prepare_dep:
ifndef HAS_DEP
	go get -u -v -d github.com/golang/dep/cmd/dep && \
	go install -v github.com/golang/dep/cmd/dep
endif

