
.PHONY: test
test: lint sqlite ## Run tests and create coverage report
	go test -race -short -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

sqlite:
	go get -u $(FLAGS) github.com/mattn/go-sqlite3
	go install $(FLAGS) github.com/mattn/go-sqlite3

.PHONY: lint
lint:
	@for file in $$(find . -name 'vendor' -prune -o -type f -name '*.go'); do \
		golint $$file; done

.PHONY: clean
clean: ## Clean up temp files and binaries
	rm -rf coverage*

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) |sort \
		|awk 'BEGIN{FS=":.*?## "};{printf "\033[36m%-30s\033[0m %s\n",$$1,$$2}'
