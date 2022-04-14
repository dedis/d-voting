versionFlag="github.com/dedis/d-voting.Version=$(shell git describe --tags)"
timeFlag="github.com/dedis/d-voting.BuildTime=$(shell date +'%d/%m/%y_%H:%M')"

lint:
	# Coding style static check.
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go mod tidy
	staticcheck ./...
#	golint -set_exit_status ./...

vet:
	@echo "⚠️ Warning: the following only works with go >= 1.14" && \
	go install ./internal/mcheck && \
	go vet -vettool=`go env GOPATH`/bin/mcheck -commentLen -ifInit ./...

# target to run all the possible checks except integration; it's a good habit to
# run it before pushing code
check: lint vet
	go test `go list ./... | grep -v /integration`

test_integration:
	go test ./integration

build:
	go build -ldflags="-X $(versionFlag) -X $(timeFlag)" ./cli/memcoin

deb:
	GOOS=linux GOARCH=amd64 make build
	cd deb-package; ./build-deb.sh; cd ..