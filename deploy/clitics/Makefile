-include Makefile.options
#####################################################################################
dist_dir=$(CURDIR)/data
executable_name=clitics
main_dir=../../cmd
service=airenas/clitics
version=0.1
commit_count=$(shell git rev-list --count HEAD)
go_build_cmd=CGO_ENABLED=0 go build -ldflags "-X main.version=$(version).$(commit_count)" 
#####################################################################################
$(dist_dir):
	mkdir -p $@
$(dist_dir)/$(executable_name): | $(dist_dir)
	cd $(main_dir)/$(executable_name) && $(go_build_cmd) -o $@
$(dist_dir)/clitics.txt: $(main_dir)/clitics/clitics.txt | $(dist_dir)
	cp $< $@
$(dist_dir)/phrases.txt: $(main_dir)/clitics/phrases.txt | $(dist_dir)
	cp $< $@	 	
#####################################################################################
build: $(dist_dir)/$(executable_name) $(dist_dir)/clitics.txt $(dist_dir)/phrases.txt
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
