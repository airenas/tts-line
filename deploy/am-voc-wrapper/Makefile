-include Makefile.options
#####################################################################################
ACOUSTICMODEL_URL?=http://localhost:8000
VOCODER_URL?=http://localhost:8001
PORT?=8002
#####################################################################################
dist_dir=$(CURDIR)/bin
executable_name=am-voc-wrapper
main_dir=../../cmd/$(executable_name)
commit_count=$(shell git rev-list --count HEAD)
version=0.5
go_build_cmd=CGO_ENABLED=0 go build -ldflags "-X main.version=$(version).$(commit_count)"
########### SERVICE#################################################################
logs:
	mkdir -p $@
install-service: am-voc.service | logs
	cp am-voc.service /etc/systemd/system/
	systemctl enable am-voc.service
uninstall-service:
	systemctl disable am-voc.service
	rm -f /etc/systemd/system/am-voc.service
am-voc.service: am-voc.service.in
	cat $< | envsubst > $@
run-service:
	$(dist_dir)/$(executable_name)
#####################################################################################
$(dist_dir):
	mkdir -p $@
$(dist_dir)/$(executable_name): | $(dist_dir)
	cd $(main_dir) && $(go_build_cmd) -o $(dist_dir)/$(executable_name) 
$(dist_dir)/config.yaml: config.yaml.in Makefile.options
	cat $< | envsubst > $@	
$(dist_dir)/$(executable_name)-$(version).$(commit_count).gz: $(dist_dir)/$(executable_name)
	gzip -c $< > $@	
#####################################################################################
build: $(dist_dir)/$(executable_name)-$(version).$(commit_count).gz
share: $(dist_dir)/$(executable_name)-$(version).$(commit_count).gz
	scp $(dist_dir)/$(executable_name)-$(version).$(commit_count).gz $(component-share)
#####################################################################################
clean:
	rm -rf $(dist_dir)

.PHONY:
	clean build dbuild dpush
.EXPORT_ALL_VARIABLES:	
