-include ../../version 
#####################################################################################
service=airenas/tts-line
version=$(tts_version)
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	cd ../../ && docker build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/tts-line/Dockerfile .

dpush: dbuild
	docker push $(tag)
#####################################################################################
.PHONY: dbuild dpush
