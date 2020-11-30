#####################################################################################
dist_dir=$(CURDIR)/data
executable_name=tts-line
main_dir=../../cmd/$(executable_name)
service=airenas/tts-line
version=0.2
go_build_cmd=CGO_ENABLED=0 go build -installsuffix -cgo
commit_count=$(shell git rev-list --count HEAD)
#####################################################################################
$(dist_dir):
	mkdir -p $@
$(dist_dir)/$(executable_name): | $(dist_dir)
	cd $(main_dir) && $(go_build_cmd) -o $(dist_dir)/$(executable_name) 
#####################################################################################
build: $(dist_dir)/$(executable_name)
########### DOCKER ##################################################################
tag=$(service):$(version).$(commit_count)

dbuild: $(dist_dir)/$(executable_name)
	docker build -t $(tag) ./

dpush: dbuild
	docker push $(tag)
#####################################################################################
clean:
	rm -rf $(dist_dir)

.PHONY:
	clean build dbuild dpush
