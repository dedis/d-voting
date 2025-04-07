.PHONY: lint check integration build deb

version=$(shell git describe --abbrev=0 --tags || echo '0.0.0')
versionFlag="go.dedis.ch/d-voting.Version=$(version)"
versionFile=$(shell echo $(version) | tr . _)
timeFlag="go.dedis.ch/d-voting.BuildTime=$(shell date +'%d/%m/%y_%H:%M')"

tidy:
	go mod tidy -go=1.23.8

lint: tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	golangci-lint run

# target to run all the possible checks except integration; it's a good habit to
# run it before pushing code
check: tidy
	go test `go list ./... | grep -v /integration`

integration:
	go test ./integration -v -race -timeout 60s

build:
	go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting ./cli/dvoting
#	GOOS=linux GOARCH=amd64 go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting-linux-amd64-$(versionFile) ./cli/dvoting
#	GOOS=darwin GOARCH=amd64 go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting-darwin-amd64-$(versionFile) ./cli/dvoting
#	GOOS=windows GOARCH=amd64 go build -ldflags="-X $(versionFlag) -X $(timeFlag)" -o dvoting-windows-amd64-$(versionFile) ./cli/dvoting

deb: build
	cd deb-package; ./build-deb.sh; cd ..
