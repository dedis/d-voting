version=$(shell git describe --abbrev=0 --tags || echo '0.0.0')
versionFlag="github.com/dedis/d-voting.Version=$(version)"
versionFile=$(shell echo $(version) | tr . _)
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
	go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting ./cli/dvoting
	GOOS=linux GOARCH=amd64 go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting-linux-amd64-$(versionFile) ./cli/dvoting
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting-darwin-amd64-$(versionFile) ./cli/dvoting
	GOOS=windows GOARCH=amd64 go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting-windows-amd64-$(versionFile) ./cli/dvoting

deb: build
	cd deb-package; ./build-deb.sh; cd ..