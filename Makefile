GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

.PHONY: test install clean all

all: maws

maws: *.go cmd/maws/*.go go.*
	cd cmd/maws && go build -o ../../maws -ldflags "-s -w -X main.version=${GIT_VER} -X main.buildDate=${DATE}" -gcflags="-trimpath=${PWD}"

install: maws
	install maws ${GOPATH}/bin

test:
	go test -race ./...

clean:
	rm -f maws
