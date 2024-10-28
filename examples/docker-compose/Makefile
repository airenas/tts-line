-include ../../version
-include Makefile.options
#####################################################################################
start: .env 
	docker compose up -d

stop:
	docker compose down

logs:
	docker compose logs	

clean: stop
#####################################################################################
.env: .env.in Makefile.options
	cat $< | envsubst > $@
#####################################################################################
test:
	curl -i -X POST http://localhost:$(service_port)/synthesize -H 'Content-type: application/json' -d '{"text": "laba diena", "speed": 0.75}'


test/ssml:
	curl -i -X POST http://localhost:$(service_port)/synthesize -H 'Content-type: application/json' -d '{"text": "<speak>laba diena</speak>"}'	

install/data/ml-tagger:
	mkdir -p ./ml-tagger
	scp -r airenas@$(c8):/home/airenas/service_data/tagger/models/* ./ml-tagger/	

.EXPORT_ALL_VARIABLES:
	

