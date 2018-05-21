GOCMD=go
GOBUILD=$(GOCMD) build

.PHONY: build
build: master worker

.PHONY: master
master:
	$(GOBUILD) -o dist/master cmd/master/main.go

.PHONY: worker
worker:
	$(GOBUILD) -o dist/worker cmd/worker/main.go

.PHONY: test
test:
	scripts/run_tests.sh

.PHONY: unittest
unittest:
	go test -v ./...

.PHONY: demo
demo:
	scripts/run_demo.sh

.PHONY: clean
clean:
	rm -rf dist