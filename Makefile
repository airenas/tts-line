-include Makefile.options
-include version
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
## run integration tests - start services, do tests, clean services
test/integration:
	cd testing/integration && $(MAKE) start test/integration clean || ( $(MAKE) clean; exit 1; ) 	
.PHONY: test/integration
#####################################################################################
## code vet and lint
test/lint: 
	go vet ./...
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run -v ./...
.PHONY: test/lint
#####################################################################################
## build docker image
docker/%/build:
	cd build/$* && $(MAKE) dbuild	
## scan docker image for vulnerabilities
docker/%/scan:
	cd build/$* && $(MAKE) dscan	
## builds all docker containers
docker/build: docker/tts-line/build docker/tts-text-clean/build	docker/acronyms/build	
.PHONY: docker/build	
#####################################################################################
## build and push docker image
docker/%/push:
	cd build/$* && $(MAKE) dpush
#####################################################################################
## creates new git tag and pushes to github 
git/create/version:
	git tag $(version)
	git push origin $(version)
.PHONY: git/create/version	
#####################################################################################
## generate diagrams
generate/diagram:
	cd info && $(MAKE) generate
.PHONY: generate/diagram
#####################################################################################
## cleans prepared data for dockeriimage generation
clean:
	go mod tidy -compat=1.19
	go clean
	cd build/acronyms && $(MAKE) clean
	cd build/clitics && $(MAKE) clean
