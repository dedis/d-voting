lint:
	# Coding style static check.
	@go get -v honnef.co/go/tools/cmd/staticcheck
	@go mod tidy
	staticcheck ./...

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