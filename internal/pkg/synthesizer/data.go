package synthesizer

//TTSData working data for one request
type TTSData struct {
	OriginalText    string
	Text            string
	TextWithNumbers string
	Audio           string
	Words           []*ProcessedWord
}

//ProcessedWord keeps one word info
type ProcessedWord struct {
	Tagged TaggedWord
}

//TaggedWord - tagger's result
type TaggedWord struct {
	Type   string
	String string
	Mi     string
	Lemma  string
}
