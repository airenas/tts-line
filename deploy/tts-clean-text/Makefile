-include ../../version 
#####################################################################################
dist_dir=$(CURDIR)/data
executable_name=tts-text-clean
main_dir=../../cmd
service=airenas/tts-text-clean
version=$(clean-version)
go_build_cmd=CGO_ENABLED=0 go build -installsuffix -cgo -ldflags "-X main.version=$(version)" 
#####################################################################################
$(dist_dir):
	mkdir -p $@
$(dist_dir)/$(executable_name): | $(dist_dir)
	cd $(main_dir)/$(executable_name) && $(go_build_cmd) -o $@
#####################################################################################
build: $(dist_dir)/$(executable_name)
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild: $(dist_dir)/$(executable_name)
	docker build -t $(tag) --build-arg BUILD_VERSION=$(version) ./

dpush: dbuild
	docker push $(tag)
#####################################################################################
clean:
	rm -rf $(dist_dir)

.PHONY:
	clean build dbuild dpush
