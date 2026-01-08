package api

type Mode string

const (
	ModeNone             Mode = ""
	ModeCharactersAsWord Mode = "charactersAsWord" // read as word, accented the last character,
	ModeCharacters       Mode = "characters"
	ModeAllAsCharacters  Mode = "all" // read all puntuations also
)

// WordInput is input structure
type WordInput struct {
	Word string `json:"word,omitempty"`
	MI   string `json:"mi,omitempty"`
	ID   string `json:"id,omitempty"`
	Mode Mode   `json:"mode,omitempty"`
}

// WordOutput is output result
type WordOutput struct {
	ID    string       `json:"id,omitempty"`
	Words []ResultWord `json:"words,omitempty"`
}

// ResultWord is output
type ResultWord struct {
	Word      string `json:"word,omitempty"`
	WordTrans string `json:"wordTrans,omitempty"`
	Syll      string `json:"syll,omitempty"`
	UserTrans string `json:"userTrans,omitempty"`
}
