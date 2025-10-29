##############################################################
service=airenas/tts-text-clean
version?=dev
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	cd ../../ && docker buildx build -t $(tag) --build-arg BUILD_VERSION=$(version) \
	-f build/tts-text-clean/Dockerfile . --load

dpush: dbuild
	docker push $(tag)
	
.PHONY: dbuild dpush
