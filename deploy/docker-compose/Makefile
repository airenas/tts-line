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
	curl -i -X POST http://localhost:$(service_port)/synthesize -H 'Content-type: application/json' -d '{"text": "laba diena", "speed": 0.75, "voice": "astra"}'


test/ssml:
	curl -i -X POST http://localhost:$(service_port)/synthesize -H 'Content-type: application/json' -d '{"text": "<speak>laba diena</speak>"}'	

#####################################################################################
file?=ssml.xml
.data:
	mkdir -p $@

.data/input.json: $(file) | .data
	@jq -n --arg text "$$(cat $(file))" --arg speed "1" --arg voice "astra" \
		'{text: $$text, speed: ($$speed | tonumber), voice: $$voice, symbolMode: "", selectedSymbols: ["(", ")", "|", "?", ","]}' > $@
.data/response.json: .data/input.json
	curl -X POST http://localhost:$(service_port)/synthesize -H 'Content-type: application/json' -d @$< -o $@
.data/output.mp3: .data/response.json	
	@jq -r '.audioAsString' .data/response.json | base64 --decode > $@

generate: .data/output.mp3
.PHONY: generate
play: .data/output.mp3
	nvlc --play-and-exit $<	
.PHONY: play
clean/data:
	rm -rf .data
.PHONY: clean/data	

#####################################################################################

install/data/ml-tagger:
	mkdir -p ./ml-tagger
	scp -r airenas@$(c8):/home/airenas/service_data/tagger/models/* ./ml-tagger/	

.EXPORT_ALL_VARIABLES:
	

