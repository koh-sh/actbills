.PHONY: setup test fmt cov tidy run lint dockerbuild dockerrun

COVFILE = coverage.out
COVHTML = cover.html

setup:
	go install github.com/mfridman/tparse@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/spf13/cobra-cli@latest

test:
	export -n GITHUB_TOKEN; export -n GITHUB_REPOSITORY; go test ./... -json | tparse -all

fmt:
	gofumpt -l -w *.go

cov:
	export -n GITHUB_TOKEN; export -n GITHUB_REPOSITORY; go test -cover ./... -coverprofile=$(COVFILE)
	go tool cover -html=$(COVFILE) -o $(COVHTML)
	rm $(COVFILE)

tidy:
	go mod tidy -v

lint:
	golangci-lint run -v

# for testing
dockerbuild:
	docker build . -t actbills:latest

dockerrun:
	docker run -it --rm -e GITHUB_TOKEN actbills:latest -v
