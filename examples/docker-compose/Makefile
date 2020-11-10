-include .env
#####################################################################################
start: 
	docker-compose up -d

stop:
	docker-compose stop && docker-compose rm -f

logs:
	docker-compose logs	

clean: stop
#####################################################################################
test:
	curl -i -X POST http://localhost:$(SERVICE_PORT)/synthesize -d '{"text": "laba diena"}'

