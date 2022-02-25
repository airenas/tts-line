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
	go test -v -race -count 1 ./...
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
docker/tts-line/build:
	cd deploy/tts-line && $(MAKE) dbuild	
.PHONY: docker/tts-line/build
## build tts-line docker image
docker/tts-clean/build:
	cd deploy/tts-clean-text && $(MAKE) dbuild	
.PHONY: docker/tts-clean/build
#####################################################################################
## build and push tts-line docker image
docker/tts-line/push:
	cd deploy/tts-line && $(MAKE) dpush
.PHONY: docker/tts-line/push
## build and push tts-clean-text docker image
docker/tts-clean/push:
	cd deploy/tts-clean-text && $(MAKE) clean dpush
.PHONY: docker/tts-clean/push

generate-diagram:
	cd info && $(MAKE) generate
#####################################################################################
## cleans prepared data for dockeriimage generation
clean:
	go mod tidy
	go clean
	cd deploy/acronyms && $(MAKE) clean
	cd deploy/clitics && $(MAKE) clean
	cd deploy/tts-clean-text && $(MAKE) clean
