generate: tts.png tts.custom.png

tts.png: tts.synthesize.puml 
	cat $< | docker run --rm -i think/plantuml -tpng > $@_
	mv $@_ $@

tts.custom.png: tts.synthesize.custom.puml 
	cat $< | docker run --rm -i think/plantuml -tpng > $@_
	mv $@_ $@
