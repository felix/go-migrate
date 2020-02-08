
.PHONY: test
test: lint
	go test -short -coverprofile=coverage.txt -covermode=atomic ./... \
		&& go tool cover -html=coverage.txt -o coverage.html

.PHONY: lint
lint:
	go vet ./...

.PHONY: clean
clean:
	rm -rf coverage*
