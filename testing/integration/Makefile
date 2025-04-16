-include Makefile.options.local
version?=dev
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' | sort
.PHONY: help
#####################################################################################
## start all containers
start: 
	docker compose up -d --build tagger lex semantika tts-line text-clean \
		acronyms accenter clitics transcriber audioconverter-rs mongo normalizer \
		comparator transliterator
.PHONY: start
## start tts-line container
start/tts-line: 
	docker compose up tts-line
.PHONY: start/tts-line
## invoke integration tests
test/integration: start 
	docker compose up --build --exit-code-from integration-tests integration-tests
.PHONY: test/integration
## invoke unit tests
test/unit:  
	docker compose up --build --exit-code-from unit-tests unit-tests
.PHONY: test/unit
## invoke tests with compose up 
test/up:  
	docker compose up tts-line integration-tests
.PHONY: test/unit
## clean everything, stops docker containers and removes them
clean:
	docker compose down
.PHONY: clean	
#####################################################################################

.EXPORT_ALL_VARIABLES:
