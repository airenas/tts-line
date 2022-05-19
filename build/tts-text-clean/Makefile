-include ../../version 
#####################################################################################
service=airenas/tts-text-clean
version=$(clean_version)
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	cd ../../ && docker build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/tts-text-clean/Dockerfile .

dpush: dbuild
	docker push $(tag)
	
.PHONY: dbuild dpush
