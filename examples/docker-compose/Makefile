-include Makefile.options
#####################################################################################
start: .env 
	docker-compose up -d

stop:
	docker-compose stop && docker-compose rm -f

logs:
	docker-compose logs	

clean: stop
#####################################################################################
.env: .env.in Makefile.options
	cat $< | envsubst > $@
#####################################################################################
test:
	curl -i -X POST http://localhost:$(SERVICE_PORT)/synthesize -d '{"text": "laba diena"}'

