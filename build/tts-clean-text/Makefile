-include ../../version 
#####################################################################################
service=airenas/tts-text-clean
version=$(clean-version)
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	cd ../../ && docker build -t $(tag) --build-arg BUILD_VERSION=$(version) -f deploy/tts-clean-text/Dockerfile .

dpush: dbuild
	docker push $(tag)
	
.PHONY: dbuild dpush
