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
	go install golang.org/x/lint/golint@latest
	golint -set_exit_status ./...
.PHONY: test/lint
#####################################################################################
## generate mock objects for test
generate/mocks: 
	go install github.com/petergtz/pegomock/...@latest
	go generate ./...
.PHONY: generate/mocks	
#####################################################################################
## build tts-line docker image
docker/tts-line/build:
	cd build/tts-line && $(MAKE) dbuild	
.PHONY: docker/tts-line/build
## build tts-line docker image
docker/tts-clean/build:
	cd build/tts-clean-text && $(MAKE) dbuild	
.PHONY: docker/tts-clean/build
## scan tts-line for vulnerabilities
docker/tts-line/scan:
	cd build/tts-line && $(MAKE) dscan	
.PHONY: docker/tts-line/build
#####################################################################################
## build and push tts-line docker image
docker/tts-line/push:
	cd build/tts-line && $(MAKE) dpush
.PHONY: docker/tts-line/push
## build and push tts-clean-text docker image
docker/tts-clean/push:
	cd build/tts-clean-text && $(MAKE) dpush
.PHONY: docker/tts-clean/push
#####################################################################################
## generate diagrams
generate/diagram:
	cd info && $(MAKE) generate
.PHONY: generate/diagram
#####################################################################################
## cleans prepared data for dockeriimage generation
clean:
	go mod tidy
	go clean
	cd build/acronyms && $(MAKE) clean
	cd build/clitics && $(MAKE) clean
