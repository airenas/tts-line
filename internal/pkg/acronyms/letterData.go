package acronyms

var letters map[string]*ldata

type ldata struct {
	ch       string
	chAccent string
	letter   string
	newWord  bool
}

func init() {
	letters = make(map[string]*ldata)
	letters["a"] = makeLD("ą3")
	letters["ą"] = makeLD("ą3")
	letters["b"] = makeLD("bė3")
	letters["c"] = makeLD("cė3")
	letters["č"] = makeLD("čė3")
	letters["d"] = makeLD("dė3")
	letters["e"] = makeLD("ę3")
	letters["ę"] = makeLD("ę3")
	letters["ė"] = makeLD("ė3")
	letters["f"] = makeLD("e4f")
	letters["g"] = makeLD("gė3")
	letters["h"] = makeLD("hą3")
	letters["i"] = makeLD("į3")
	letters["į"] = makeLD("į3")
	letters["y"] = makeLD("į3")
	letters["j"] = makeLD("jO4t")
	letters["k"] = makeLD("ką3")
	letters["l"] = makeLD("el3")
	letters["m"] = makeLD("em3")
	letters["n"] = makeLD("en3")
	letters["o"] = makeLD("o3")
	letters["p"] = makeLD("pė3")
	letters["q"] = makeLD("kų3")
	letters["r"] = makeLD("er3")
	letters["s"] = makeLD("e4s")
	letters["š"] = makeLD("e4š")
	letters["t"] = makeLD("tė3")
	letters["u"] = makeLD("ų3")
	letters["ū"] = makeLD("ų3")
	letters["ų"] = makeLD("ų3")
	letters["v"] = makeLD("vė3")
	wl := makeLD("da4bl-vė")
	wl.newWord = true
	letters["w"] = wl
	letters["x"] = makeLD("i4ks")
	letters["z"] = makeLD("zė3")
	letters["ž"] = makeLD("žė3")
	wl = makeLD("ta3-škas")
	wl.newWord = true
	letters["."] = wl
	for k, v := range letters {
		v.letter = k
	}
}

func makeLD(ch string) *ldata {
	var r ldata
	r.chAccent = ch
	r.ch = trimAccent(ch)
	return &r
}
