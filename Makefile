-include Makefile.options

test: 
	go test -v ./...

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

clean:
	cd deploy/tts-line && $(MAKE) clean

