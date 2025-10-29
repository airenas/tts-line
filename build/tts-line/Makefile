#####################################################################################
service=airenas/tts-line
version?=dev
debug?=
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	docker buildx build -t $(tag) --build-arg BUILD_VERSION=$(version) \
		-f Dockerfile$(debug) ../../ --load

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)	
#####################################################################################
.PHONY: dbuild dpush
