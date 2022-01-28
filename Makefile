-include Makefile.options
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' | sort 
.PHONY: help
#####################################################################################
## call units tests
test/unit: 
	go test -v -race ./...
.PHONY: test/unit
#####################################################################################
## code vet and lint
test/lint: 
	go vet ./...
	go get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
.PHONY: test/lint
#####################################################################################
generate: 
	go get github.com/petergtz/pegomock/...
	go generate ./...

build:
	cd cmd/tts-line/ && go build .

run:
	cd cmd/tts-line/ && go run . -c config.yml	

build-docker:
	cd deploy/tts-line && $(MAKE) dbuild	

push-docker:
	cd deploy/tts-line && $(MAKE) dpush

generate-diagram:
	cd info && $(MAKE) generate

clean:
	cd deploy/tts-line && $(MAKE) clean

