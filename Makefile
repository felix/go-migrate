
test: ## Run tests and create coverage report
	go test -short -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

sqlite:
	go get -u $(FLAGS) github.com/mattn/go-sqlite3
	go install $(FLAGS) github.com/mattn/go-sqlite3

lint:
	@for file in $$(find . -name 'vendor' -prune -o -type f -name '*.go'); do \
		golint $$file; done

clean: ## Clean up temp files and binaries
	rm -rf coverage* vendor Gopkg*

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) |sort \
		|awk 'BEGIN{FS=":.*?## "};{printf "\033[36m%-30s\033[0m %s\n",$$1,$$2}'

.PHONY: help test lint clean
