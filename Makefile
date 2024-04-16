
.DEFAULT_GOAL := all

all: go.mod cmd/smf/*go
	go mod tidy
	GOBIN=`realpath ../export/bin` go install ./...

local: go.mod cmd/smf/*go
	go mod tidy
	go fmt ./...
	go vet ./...
	- staticcheck ./...
	mkdir -p ./bin
	GOBIN=`realpath ./bin` go install ./...

test: local
	./test.sh
	./run.sh

sgwpgw: local
	bin/sgwpgw 127.0.0.2:8805 127.0.0.3:8805 &
	bin/smf -peer 127.0.0.2:8805 -sgw 127.0.0.3:8805
	killall sgwpgw
