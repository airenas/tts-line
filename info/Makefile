generate: tts.png

tts.png: tts.synthesize.puml 
	cat $< | docker run --rm -i think/plantuml -tpng > $@_
	mv $@_ $@
