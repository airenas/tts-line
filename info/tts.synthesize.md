```mermaid
sequenceDiagram
    %% autonumber
    title TTS Synthesize Method

    actor User
    participant tts as TTS
    participant DB as DB
    participant Clean
    participant Normalizer
    participant num as NumberReplace
    participant Tagger
    participant Lex
    participant Morf
    participant Transliterator
    participant obscene as Obscene Filter
    participant Acronym
    participant Accenter
    participant Clitic
    participant Transcriber

    box GPU 
        participant amVoc as AMVocWrapper
        participant AM
        participant Vocoder
    end

    participant converter as Mp3 M4a converter

    User ->> tts: synthesize
    activate tts

    tts->>tts: check for empty/too large text

    tts ->>+ DB: save original text
    DB -->>- tts: 

    tts->>+Clean: clean text, drop html tags
    Clean -->>- tts: 

    tts ->>+ DB: save cleaned text
    DB -->>- tts: 

    tts ->>+ Normalizer: normalize text, change some numbers
    Normalizer -->>- tts: 

    tts ->>+ num: 
    num -->>- tts: 


    tts ->>+ Tagger: 
    Tagger ->>+ Lex: 
    Lex -->>- Tagger: 

    Tagger ->>+ Morf: 
    Morf -->>- Tagger: 

    Tagger -->>- tts: 

    tts ->>+ Normalizer: replace URLs to words
    Normalizer -->>- tts: 

    tts ->>+ Tagger: Tag words from URLs
    Tagger -->>- tts: 

    tts ->>+ Transliterator: change non LT words as readable ones 
    Transliterator -->>- tts: 

    tts ->>+ DB: save normalized text
    DB -->>- tts: 

    tts ->> tts: Do minimal NER

    tts ->> tts: Split into batches

    par Parallel Processing for each batch
        tts ->>+ obscene: 
        obscene -->>- tts: 

        tts ->>+ Acronym: 
        Acronym -->>- tts: 

        tts ->>+ Accenter: 
        Accenter -->>- tts: 

        tts ->>+ Clitic: 
        Clitic -->>- tts: 

        tts ->>+ Transcriber: 
        Transcriber -->>- tts: 

        tts ->>+ amVoc: 
        amVoc ->>+ AM: 
        AM -->>- amVoc: 

        amVoc ->>+ Vocoder: 
        Vocoder -->>- amVoc: 
        amVoc -->>- tts: 
    end

    tts ->> tts: Calculate batch volume/Unify volume
    tts ->> tts: Join audio batches

    
    tts ->>+ converter: 
    converter -->>- tts: 

    tts -->>- User: response
```