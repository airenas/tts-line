package api

// WordInput is input structure
type WordInput struct {
	Word string `json:"word,omitempty"`
	MI   string `json:"mi,omitempty"`
	ID   string `json:"id,omitempty"`
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
