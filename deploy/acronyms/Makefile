-include Makefile.options
#####################################################################################
dist_dir=$(CURDIR)/data
executable_name=acronyms
main_dir=../../cmd
data_dir?=$(main_dir)
service=airenas/acronyms
version=0.1
commit_count=$(shell git rev-list --count HEAD)
go_build_cmd=CGO_ENABLED=0 go build -ldflags "-X main.version=$(version).$(commit_count)" 
#####################################################################################
$(dist_dir):
	mkdir -p $@
$(dist_dir)/$(executable_name): | $(dist_dir)
	cd $(main_dir)/$(executable_name) && $(go_build_cmd) -o $@
$(dist_dir)/acronyms.txt: $(data_dir)/acronyms/acronyms.txt | $(dist_dir)
	cp $< $@ 	
#####################################################################################
build: $(dist_dir)/$(executable_name) $(dist_dir)/acronyms.txt
########### DOCKER ##################################################################
tag=$(service):$(version).$(commit_count)

dbuild: build
	docker build -t $(tag) --build-arg SECRET=$(secret) ./

dpush: dbuild
	docker push $(tag)
#####################################################################################
clean:
	rm -rf $(dist_dir)

.PHONY:
	clean build dbuild dpush
