-include ../../version 
#####################################################################################
dist_dir=$(CURDIR)/data
executable_name=tts-line
main_dir=../../cmd
service=airenas/tts-line
version=$(tts_version)
go_build_cmd=CGO_ENABLED=0 go build -ldflags "-X main.version=$(version)" 
#####################################################################################
$(dist_dir):
	mkdir -p $@
$(dist_dir)/$(executable_name): | $(dist_dir)
	cd $(main_dir)/$(executable_name) && $(go_build_cmd) -o $@
$(dist_dir)/tts-export: | $(dist_dir)
	cd $(main_dir)/tts-export && $(go_build_cmd) -o $@ 	
#####################################################################################
build: $(dist_dir)/$(executable_name) $(dist_dir)/tts-export
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild: $(dist_dir)/$(executable_name) $(dist_dir)/tts-export
	docker build -t $(tag) --build-arg BUILD_VERSION=$(version)  ./

dpush: dbuild
	docker push $(tag)
#####################################################################################
clean:
	rm -rf $(dist_dir)

.PHONY:
	clean build dbuild dpush
