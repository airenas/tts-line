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

#####################################################################################
## build tts-line docker image
build/tts-line:
	cd deploy/tts-line && $(MAKE) clean dbuild	
.PHONY: build/tts-line
#####################################################################################
## build and push tts-line docker image
docker/tts-line/push:
	cd deploy/tts-line && $(MAKE) clean dpush
.PHONY: docker/tts-line/push

generate-diagram:
	cd info && $(MAKE) generate
#####################################################################################
## cleans prepared data for dockeriimage generation
clean:
	cd deploy/tts-line && $(MAKE) clean
	cd deploy/acronyms && $(MAKE) clean
	cd deploy/clitics && $(MAKE) clean
	cd deploy/tts-clean-text && $(MAKE) clean
