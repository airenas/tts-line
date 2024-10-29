package api

// CliticsInput is input structure
type CliticsInput struct {
	Type   string `json:"type,omitempty"`
	String string `json:"string,omitempty"`
	Mi     string `json:"mi,omitempty"`
	Lemma  string `json:"lemma,omitempty"`
	ID     int    `json:"id,omitempty"`
}

// CliticsOutput is output result
type CliticsOutput struct {
	ID         int    `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	AccentType string `json:"accentType,omitempty"`
	Accent     int    `json:"accent,omitempty"`
	Pos        int    `json:"pos,omitempty"`
}

const (
	//TypeStatic must static accent be used
	TypeStatic = "STATIC"
	//TypeNone indicates no accent for word
	TypeNone = "NONE"
	//TypeNormal indicates no impact on the accent according clitics rule
	TypeNormal = "NORMAL"
)
