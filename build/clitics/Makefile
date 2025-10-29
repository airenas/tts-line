-include Makefile.options
#####################################################################################
service=airenas/clitics
version?=dev
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	docker buildx build -t $(tag) \
		--build-arg BUILD_VERSION=$(version) \
		--build-arg SECRET=$(CLITICS_SECRET) \
		--build-arg SECRET_FROM=$(CLITICS_SECRET_FROM) \
		-f Dockerfile ../../ --load

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)	
#####################################################################################
.PHONY: dbuild dpush