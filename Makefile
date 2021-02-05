GO	?= go1.16beta1

test: lint
	$(GO) test -race -short -coverprofile=coverage.txt -covermode=atomic ./... \
		&& $(GO) tool cover -func=coverage.txt

lint: ; $(GO) vet ./...

clean: ; rm -rf coverage*
