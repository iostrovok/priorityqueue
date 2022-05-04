CURDIR := $(shell pwd)
GOBIN := $(CURDIR)/bin/
ENV:=GOBIN=$(GOBIN)
DIR:=FILE_DIR=$(CURDIR)/testfiles TEST_SOURCE_PATH=$(CURDIR)
GODEBUG:=GODEBUG=gocacheverify=1
LOADENV:=GO111MODULE=on GONOSUMDB="*" GOPROXY=direct $(ENV) CURDIR=$(CURDIR)

##
## List of commands:
##

## default:
all: mod test

tests: clean-cache test

test:
	@echo "----"
	@echo "Run race test for ./"
	go test -cover -race ./

mod:
	@echo "======================================================================"
	@echo "Run MOD"
	$(LOADENV) go mod verify
	$(LOADENV) go mod tidy
	$(LOADENV) go mod vendor
	$(LOADENV) go mod download
	$(LOADENV) go mod verify

clean-cache:
	@echo "clean-cache started..."
	go clean -cache
	go clean -testcache
	@echo "clean-cache complete!"

clean:
	@echo "clean started..."
	rm -rf ./vendor
	go clean -cache
	go clean -testcache
	@echo "clean complete!"
