LDFLAGS += -X "main.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
#LDFLAGS += -X "main.BuildGitHash=$(shell git rev-parse HEAD)"
NOW = $(shell date -u '+%Y%m%d%I%M%S')

TAGS = ""
BUILD_FLAGS = "-v"

all: build build-win64

vendoring:
	go mod tidy

build: vendoring
	CGO_ENABLED=1 go build -o bin/checkserver $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' github.com/dafanasiev/OTPCredentialProvider-backend/app/checkserver
	CGO_ENABLED=1 go build -o bin/showqr $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' github.com/dafanasiev/OTPCredentialProvider-backend/app/showqr

build-win64: vendoring
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o bin/checkserver.exe $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' github.com/dafanasiev/OTPCredentialProvider-backend/app/checkserver
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o bin/showqr.exe $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' github.com/dafanasiev/OTPCredentialProvider-backend/app/showqr

clean:
	rm -rf bin/

