-include Makefile.options
#####################################################################################
ACOUSTICMODEL_URL?=http://localhost:8002/model
VOCODER_URL?=http://localhost:8001/model
PORT?=8000
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
dist_dir=$(CURDIR)/bin
executable_name=am-voc-wrapper
main_dir=../../cmd/$(executable_name)
go_build_cmd=CGO_ENABLED=0 go build -installsuffix -cgo
#####################################################################################
$(dist_dir):
	mkdir -p $@
$(dist_dir)/$(executable_name): | $(dist_dir)
	cd $(main_dir) && $(go_build_cmd) -o $(dist_dir)/$(executable_name) 
$(dist_dir)/config.yaml: config.yaml.in Makefile.options
	cat $< | envsubst > $@	
#####################################################################################
build: $(dist_dir)/$(executable_name) $(dist_dir)/config.yaml
#####################################################################################
clean:
	rm -rf $(dist_dir)

.PHONY:
	clean build dbuild dpush
.EXPORT_ALL_VARIABLES:	
