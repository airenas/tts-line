-include Makefile.options 
#####################################################################################
service=airenas/acronyms
version?=dev
########### DOCKER ##################################################################
tag=$(service):$(version)

data:
	mkdir -p $@

# data/acronyms.txt: $(acronyms_data_dir)/acronyms.txt | data
# 	cp $^ $@ 

# dbuild: data/acronyms.txt
dbuild: 
	docker buildx build -t $(tag) --build-arg BUILD_VERSION=$(version) \
		-f Dockerfile$(debug) --build-arg ACRONYMS_SECRET=$(ACRONYMS_SECRET) --build-arg ACRONYMS_SECRET_FROM=$(ACRONYMS_SECRET_FROM) ../../ --load

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)	
#####################################################################################

#####################################################################################
clean:
	rm -rf $(dist_dir)
.PHONY: clean dbuild dpush